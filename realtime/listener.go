package rts

import (
	"context"
	"strconv"
	"time"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/rdb"
	"github.com/go-redis/redis/v8"
	"github.com/rabbitmq/amqp091-go"
)

func (s *RTS) notificationListener() {
	for {
		select {
		case <-s.amqph.OnDataPolicyDeleted():
			for _, cp := range s.pulling {
				cp.Close()
			}
		case n := <-s.amqph.OnContainerUpdated():
			c, ok := s.pulling[n.Base.Id]
			if !ok {
				continue
			}
			c.Close()
		case id := <-s.amqph.OnContainerDeleted():
			c, ok := s.pulling[id]
			if !ok {
				continue
			}
			c.Close()
		case n := <-s.amqph.OnMetricUpdated():
			c, ok := s.pulling[n.Base.ContainerId]
			if !ok {
				continue
			}
			m, ok := c.Metrics[n.Base.Id]
			if !ok {
				continue
			}
			m.Stop()
		case pair := <-s.amqph.OnMetricDeleted():
			cp, ok := s.pulling[pair.ContainerId]
			if !ok {
				continue
			}
			m, ok := cp.Metrics[pair.Id]
			if !ok {
				continue
			}
			m.Stop()
		case <-s.Done():
			return
		}
	}
}

func (s *RTS) metricDataRequestListener() {
	msgs, err := s.amqph.Listen(amqp.QueueRTSMetricDataRequest, amqp.ExchangeRTSMetricDataRequest)
	if err != nil {
		s.log.Panic("Fail to listen amqp messages", logger.ErrField(err))
		return
	}
	for {
		select {
		case d := <-msgs:
			ctx := context.Background()

			if len(d.CorrelationId) == 0 {
				s.log.Warn("Received a rts data request but no correlation id was provided")
				continue
			}

			var r models.MetricRequest
			err = amqp.Decode(d.Body, &r)
			if err != nil {
				s.log.Error("Fail to decode message body", logger.ErrField(err))
				continue
			}

			metricIdString := strconv.FormatInt(int64(r.ContainerId), 10)
			s.log.Debug("Get metric data request received, container id: " + metricIdString)

			bytes, err := s.cache.Get(ctx, rdb.CacheMetricDataKey(r.MetricId))
			if err != nil {
				if err != redis.Nil {
					s.log.Error("Fail to get metric data on redis", logger.ErrField(err))
					continue
				}

				routingKey, err := amqp.GetDataRoutingKey(r.ContainerType)

				// if there is no routing key for this container type,
				// send response with nil value
				if err != nil {
					b, err := amqp.Encode(models.MetricDataResponse{
						MetricBasicDataReponse: models.MetricBasicDataReponse{
							Id:           r.MetricId,
							Type:         r.MetricType,
							Value:        nil,
							DataPolicyId: r.DataPolicyId,
							Failed:       false,
						},
						ContainerId: r.ContainerId,
					})
					if err != nil {
						s.log.Error("Fail to encode amqp message", logger.ErrField(err))
						continue
					}

					s.publishRTSMetricData(amqp091.Publishing{
						Expiration:    amqp.DefaultExp,
						Body:          b,
						Type:          amqp.FromMessageType(amqp.OK),
						CorrelationId: d.CorrelationId,
					})
					continue
				}

				config, err := s.getRTSMetricConfig(ctx, r.MetricId)
				if err != nil {
					s.log.Error("Fail to get metric rts information", logger.ErrField(err))
					continue
				}

				s.startMetricPulling(r, config)
				go func(correlationId string, r models.MetricRequest) {
					// send metric data request to translators
					s.amqph.PublisherCh <- models.DetailedPublishing{
						Exchange:   amqp.ExchangeMetricDataRequest,
						RoutingKey: routingKey,
						Publishing: amqp091.Publishing{
							Expiration:    amqp.DefaultExp,
							Headers:       amqp.RouteHeader("rts"),
							CorrelationId: correlationId,
							Body:          d.Body,
						},
					}
					s.pendingMetricDataRequest[correlationId] = config
					defer func(correlationId string) {
						delete(s.pendingMetricDataRequest, correlationId)
					}(correlationId)

					res, err := s.plumber.Listen(correlationId, time.Second*25)
					if err != nil {
						s.log.Warn("Plumber timeout, no data response was available")
						return
					}

					s.publishRTSMetricData(amqp091.Publishing{
						Type:          res.Type,
						Body:          res.Body,
						CorrelationId: correlationId,
					})
				}(d.CorrelationId, r)
				continue
			}

			s.restartMetricPulling(r)
			s.publishRTSMetricData(amqp091.Publishing{
				Expiration:    amqp.DefaultExp,
				Body:          bytes,
				Type:          amqp.FromMessageType(amqp.OK),
				CorrelationId: d.CorrelationId,
			})
			s.log.Debug("Metric data fetched on cache, metric id: " + metricIdString)
		case <-s.Done():
			return
		}
	}
}

func (s *RTS) metricDataListener() {
	msgs, err := s.amqph.Listen(amqp.QueueRTSMetricDataResponse, amqp.ExchangeMetricDataResponse,
		models.ListenerOptions{
			Bind: models.QueueBindOptions{
				RoutingKey: "rts",
			},
		},
	)
	if err != nil {
		s.log.Panic("Fail to listen amqp msgs", logger.ErrField(err))
		return
	}
	for {
		select {
		case d := <-msgs:
			s.plumber.Send(d)
			ctx := context.Background()

			if amqp.ToMessageType(d.Type) != amqp.OK {
				continue
			}

			var m models.MetricDataResponse
			err := amqp.Decode(d.Body, &m)
			if err != nil {
				s.log.Error("Fail to decode amqp message body", logger.ErrField(err))
				continue
			}

			config, ok := s.pendingMetricDataRequest[d.CorrelationId]
			if !ok {
				var err error
				config, err = s.getRTSMetricConfig(ctx, m.Id)
				if err != nil {
					s.log.Error("Fail to get metric configuration", logger.ErrField(err))
					continue
				}
			}

			err = s.cache.Set(ctx, d.Body, rdb.CacheMetricDataKey(m.Id), time.Millisecond*time.Duration(config.CacheDuration))
			if err != nil {
				s.log.Error("Fail to save metric data on cache", logger.ErrField(err))
				continue
			}

			s.log.Debug("metric data received, id: " + strconv.FormatInt(m.Id, 10))
		case <-s.Done():
			return
		}
	}
}

func (s *RTS) metricsDataListener() {
	msgs, err := s.amqph.Listen(amqp.QueueRTSMetricsDataResponse, amqp.ExchangeMetricsDataResponse,
		models.ListenerOptions{
			Bind: models.QueueBindOptions{
				RoutingKey: "rts",
			},
		},
	)
	if err != nil {
		s.log.Panic("Fail to listen amqp msgs", logger.ErrField(err))
		return
	}
	for {
		select {
		case d := <-msgs:
			ctx := context.Background()

			if amqp.ToMessageType(d.Type) != amqp.OK {
				continue
			}

			var m models.MetricsDataResponse
			err := amqp.Decode(d.Body, &m)
			if err != nil {
				s.log.Error("Fail to decode amqp message body", logger.ErrField(err))
				continue
			}

			p, ok := s.pulling[m.ContainerId]
			if !ok {
				continue
			}

			for _, v := range m.Metrics {
				if v.Failed {
					continue
				}

				mp, ok := p.Metrics[v.Id]
				if !ok {
					continue
				}

				b, err := amqp.Encode(models.MetricDataResponse{
					MetricBasicDataReponse: v,
					ContainerId:            m.ContainerId,
				})
				if err != nil {
					s.log.Error("Fail to encode metric response", logger.ErrField(err))
					continue
				}

				err = s.cache.Set(ctx, b, rdb.CacheMetricDataKey(v.Id), time.Millisecond*time.Duration(mp.CacheDuration))
				if err != nil {
					s.log.Error("Fail to save metric data on cache", logger.ErrField(err))
					continue
				}
			}

			s.log.Debug("Metrics data received, container id: " + strconv.FormatInt(int64(m.ContainerId), 10))
		case <-s.Done():
			return
		}
	}
}
