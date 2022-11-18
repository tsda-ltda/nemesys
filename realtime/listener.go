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

func (s *RTS) containerListener() {
	for {
		select {
		case n := <-s.amqph.OnContainerUpdated():
			// close container pulling on update
			c, ok := s.pulling[n.Base.Id]
			if !ok {
				return
			}
			c.Close()
		case id := <-s.amqph.OnContainerDeleted():
			c, ok := s.pulling[id]
			if !ok {
				return
			}
			c.Close()
		}
	}
}

func (s *RTS) metricListener() {
	for {
		select {
		case n := <-s.amqph.OnMetricUpdated():
			// stop metric pulling on update
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
			// stop metric pulling on metric delete
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

		// check if correlation id empty
		if len(d.CorrelationId) == 0 {
			s.log.Warn("receive a rts data request but no correlation id was provided")
			continue
		}

		// decode message body
		var r models.MetricRequest
		err = amqp.Decode(d.Body, &r)
		if err != nil {
			s.log.Error("fail to decode message body", logger.ErrField(err))
			continue
		}

		// parse metric id to string
		metricIdString := strconv.FormatInt(int64(r.ContainerId), 10)
		s.log.Debug("get metric data request received, container id: " + metricIdString)

		// get data on cache
		bytes, err := s.cache.Get(ctx, rdb.RDBCacheMetricDataKey(r.MetricId))
		if err != nil {
			if err != redis.Nil {
				s.log.Error("fail to get metric data on redis", logger.ErrField(err))
				continue
			}

			// get metric container type and RTS configuration
			res, err := s.pgConn.Metrics.GetRTSConfig(ctx, r.MetricId)
			if err != nil {
				s.log.Error("fail to get metric rts information", logger.ErrField(err))
				continue
			}

			// check if configuration does not exists
			if !res.Exists {
				s.log.Warn("fail to get metric rts information, metric does not exist")
				continue
			}

			// start pulling
			s.startMetricPulling(r, res.RTSConfig)

			// publish data when available
			go func(correlationId string, r models.MetricRequest) {
				// send data request
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

				// set pending request
				s.pendingMetricData[correlationId] = res.RTSConfig

				// delete channel
				defer func(correlationId string) {
					delete(s.pendingMetricData, correlationId)
				}(correlationId)

				// wait response
				res, err := s.plumber.Listen(correlationId, time.Second*5)
				if err != nil {
					s.log.Warn("plumber timeout, no data response was available")
					return
				}

				// publish data
				s.publishRTSMetricData(amqp091.Publishing{
					Type:          res.Type,
					Body:          res.Body,
					CorrelationId: correlationId,
				})
			}(d.CorrelationId, r)
			continue
		}

		// restart/start pulling
		s.restartMetricPulling(r)

		// publish data
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

		// check for pending request
		info, ok := s.pendingMetricData[d.CorrelationId]

		// send data as response for a possible listener
		s.plumber.Send(d)

		// check if info exists
		if !ok {
			continue
		}

		// check message type
		if amqp.ToMessageType(d.Type) != amqp.OK {
			continue
		}

		// decode message body
		var m models.MetricDataResponse
		err := amqp.Decode(d.Body, &m)
		if err != nil {
			s.log.Error("fail to decode amqp message body", logger.ErrField(err))
			continue
		}

		// save on cache
		err = s.cache.Set(ctx, d.Body, rdb.RDBCacheMetricDataKey(m.Id), time.Millisecond*time.Duration(info.CacheDuration))
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

		// check if message type
		if amqp.ToMessageType(d.Type) != amqp.OK {
			continue
		}

		// decode message body
		var m models.MetricsDataResponse
		err := amqp.Decode(d.Body, &m)
		if err != nil {
			s.log.Error("fail to decode amqp message body", logger.ErrField(err))
			continue
		}

		// get container pulling
		p, ok := s.pulling[m.ContainerId]
		if !ok {
			continue
		}

		for _, v := range m.Metrics {
			// check if metric has failed
			if v.Failed {
				continue
			}

			// get metric pulling
			mp, ok := p.Metrics[v.Id]
			if !ok {
				continue
			}

			// encode metric
			b, err := amqp.Encode(models.MetricDataResponse{
				MetricBasicDataReponse: v,
				ContainerId:            m.ContainerId,
			})
			if err != nil {
				s.log.Error("fail to encode metric response", logger.ErrField(err))
				continue
			}

			// save on cache
			err = s.cache.Set(ctx, b, rdb.RDBCacheMetricDataKey(v.Id), time.Millisecond*time.Duration(mp.CacheDuration))
			if err != nil {
				s.log.Error("fail to save metric data on cache", logger.ErrField(err))
				continue
			}
		}

		s.log.Debug("metrics data received, container id: " + strconv.FormatInt(int64(m.ContainerId), 10))
	}
}
