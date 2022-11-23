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
		}
	}
}

func (s *RTS) metricDataRequestListener() {
	msgs, err := s.amqph.Listen(amqp.QueueRTSMetricDataRequest, amqp.ExchangeRTSMetricDataRequest)
	if err != nil {
		s.log.Panic("fail to listen amqp messages", logger.ErrField(err))
		return
	}
	for d := range msgs {
		ctx := context.Background()

		if len(d.CorrelationId) == 0 {
			s.log.Warn("receive a rts data request but no correlation id was provided")
			continue
		}

		var r models.MetricRequest
		err = amqp.Decode(d.Body, &r)
		if err != nil {
			s.log.Error("fail to decode message body", logger.ErrField(err))
			continue
		}

		metricIdString := strconv.FormatInt(int64(r.ContainerId), 10)
		s.log.Debug("get metric data request received, container id: " + metricIdString)

		bytes, err := s.cache.Get(ctx, rdb.CacheMetricDataKey(r.MetricId))
		if err != nil {
			if err != redis.Nil {
				s.log.Error("fail to get metric data on redis", logger.ErrField(err))
				continue
			}

			res, err := s.pgConn.GetMetricRTSConfig(ctx, r.MetricId)
			if err != nil {
				s.log.Error("fail to get metric rts information", logger.ErrField(err))
				continue
			}
			if !res.Exists {
				s.log.Warn("fail to get metric rts information, metric does not exist")
				continue
			}

			s.startMetricPulling(r, res.RTSConfig)
			go func(correlationId string, r models.MetricRequest) {
				// send metric data request to translators
				s.amqph.PublisherCh <- models.DetailedPublishing{
					Exchange:   amqp.ExchangeMetricDataRequest,
					RoutingKey: amqp.GetDataRoutingKey(r.ContainerType),
					Publishing: amqp091.Publishing{
						Expiration:    amqp.DefaultExp,
						Headers:       amqp.RouteHeader("rts"),
						CorrelationId: correlationId,
						Body:          d.Body,
					},
				}
				s.pendingMetricData[correlationId] = res.RTSConfig
				defer func(correlationId string) {
					delete(s.pendingMetricData, correlationId)
				}(correlationId)

				res, err := s.plumber.Listen(correlationId, time.Second*25)
				if err != nil {
					s.log.Warn("plumber timeout, no data response was available")
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
		s.log.Debug("metric data fetched on cache, metric id: " + metricIdString)
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
		s.log.Panic("fail to listen amqp msgs", logger.ErrField(err))
		return
	}
	for d := range msgs {
		ctx := context.Background()

		s.plumber.Send(d)

		info, ok := s.pendingMetricData[d.CorrelationId]
		if !ok {
			continue
		}

		if amqp.ToMessageType(d.Type) != amqp.OK {
			continue
		}

		var m models.MetricDataResponse
		err := amqp.Decode(d.Body, &m)
		if err != nil {
			s.log.Error("fail to decode amqp message body", logger.ErrField(err))
			continue
		}

		err = s.cache.Set(ctx, d.Body, rdb.CacheMetricDataKey(m.Id), time.Millisecond*time.Duration(info.CacheDuration))
		if err != nil {
			s.log.Error("fail to save metric data on cache", logger.ErrField(err))
			continue
		}

		s.log.Debug("metric data received, id: " + strconv.FormatInt(m.Id, 10))
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
		s.log.Panic("fail to listen amqp msgs", logger.ErrField(err))
		return
	}
	for d := range msgs {
		ctx := context.Background()

		if amqp.ToMessageType(d.Type) != amqp.OK {
			continue
		}

		var m models.MetricsDataResponse
		err := amqp.Decode(d.Body, &m)
		if err != nil {
			s.log.Error("fail to decode amqp message body", logger.ErrField(err))
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
				s.log.Error("fail to encode metric response", logger.ErrField(err))
				continue
			}

			err = s.cache.Set(ctx, b, rdb.CacheMetricDataKey(v.Id), time.Millisecond*time.Duration(mp.CacheDuration))
			if err != nil {
				s.log.Error("fail to save metric data on cache", logger.ErrField(err))
				continue
			}
		}

		s.log.Debug("metrics data received, container id: " + strconv.FormatInt(int64(m.ContainerId), 10))
	}
}
