package rts

import (
	"context"
	"strconv"
	"time"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/amqph"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/rdb"
	"github.com/go-redis/redis/v8"
	"github.com/rabbitmq/amqp091-go"
)

func (s *RTS) metricDataRequestHandler(d amqp091.Delivery) {
	ctx := context.Background()

	if len(d.CorrelationId) == 0 {
		s.log.Warn("Received a rts data request but no correlation id was provided")
		return
	}

	var r models.MetricRequest
	err := amqp.Decode(d.Body, &r)
	if err != nil {
		s.log.Error("Fail to decode message body", logger.ErrField(err))
		return
	}

	responseRk, err := amqp.GetRoutingKeyFromHeader(d.Headers)
	if err != nil {
		s.log.Error("Fail to get response routing key from header", logger.ErrField(err))
		return
	}

	metricIdString := strconv.FormatInt(int64(r.ContainerId), 10)
	s.log.Debug("Get metric data request received, container id: " + metricIdString)

	bytes, err := s.cache.Get(ctx, rdb.CacheMetricDataKey(r.MetricId))
	if err != nil {
		if err != redis.Nil {
			s.log.Error("Fail to get metric data on redis", logger.ErrField(err))
			return
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
				return
			}

			s.publishRTSMetricData(amqp091.Publishing{
				Expiration:    amqp.DefaultExp,
				Body:          b,
				Type:          amqp.FromMessageType(amqp.OK),
				CorrelationId: d.CorrelationId,
			}, responseRk)
			return
		}

		config, err := s.getRTSMetricConfig(ctx, r.MetricId)
		if err != nil {
			s.log.Error("Fail to get metric rts information", logger.ErrField(err))
			return
		}

		s.startMetricPulling(r, config)
		go func(correlationId string, r models.MetricRequest) {
			// send metric data request to translators
			s.amqph.Publish(amqph.Publish{
				Exchange:   amqp.ExchangeMetricDataReq,
				RoutingKey: routingKey,
				Publishing: amqp091.Publishing{
					Expiration:    amqp.DefaultExp,
					Headers:       amqp.RouteHeader(s.GetServiceIdent()),
					CorrelationId: correlationId,
					Body:          d.Body,
				},
			})

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
			}, responseRk)
		}(d.CorrelationId, r)
		return
	}

	s.restartMetricPulling(r)
	s.publishRTSMetricData(amqp091.Publishing{
		Expiration:    amqp.DefaultExp,
		Body:          bytes,
		Type:          amqp.FromMessageType(amqp.OK),
		CorrelationId: d.CorrelationId,
	}, responseRk)
	s.log.Debug("Metric data fetched on cache, metric id: " + metricIdString)
}

func (s *RTS) metricDataHandler(d amqp091.Delivery) {
	s.plumber.Send(d)
	if amqp.ToMessageType(d.Type) != amqp.OK {
		return
	}
	ctx := context.Background()

	var m models.MetricDataResponse
	err := amqp.Decode(d.Body, &m)
	if err != nil {
		s.log.Error("Fail to decode amqp message body", logger.ErrField(err))
		return
	}

	config, ok := s.pendingMetricDataRequest[d.CorrelationId]
	if !ok {
		var err error
		config, err = s.getRTSMetricConfig(ctx, m.Id)
		if err != nil {
			s.log.Error("Fail to get metric configuration", logger.ErrField(err))
			return
		}
	}

	err = s.cache.Set(ctx, d.Body, rdb.CacheMetricDataKey(m.Id), time.Millisecond*time.Duration(config.CacheDuration))
	if err != nil {
		s.log.Error("Fail to save metric data on cache", logger.ErrField(err))
		return
	}

	s.log.Debug("Metric data received, id: " + strconv.FormatInt(m.Id, 10))
}

func (s *RTS) metricsDataHandler(d amqp091.Delivery) {
	if amqp.ToMessageType(d.Type) != amqp.OK {
		return
	}
	ctx := context.Background()

	var m models.MetricsDataResponse
	err := amqp.Decode(d.Body, &m)
	if err != nil {
		s.log.Error("Fail to decode amqp message body", logger.ErrField(err))
		return
	}

	p, ok := s.pulling[m.ContainerId]
	if !ok {
		return
	}

	for _, v := range m.Metrics {
		if v.Failed {
			return
		}

		mp, ok := p.Metrics[v.Id]
		if !ok {
			return
		}

		b, err := amqp.Encode(models.MetricDataResponse{
			MetricBasicDataReponse: v,
			ContainerId:            m.ContainerId,
		})
		if err != nil {
			s.log.Error("Fail to encode metric response", logger.ErrField(err))
			return
		}

		err = s.cache.Set(ctx, b, rdb.CacheMetricDataKey(v.Id), time.Millisecond*time.Duration(mp.CacheDuration))
		if err != nil {
			s.log.Error("Fail to save metric data on cache", logger.ErrField(err))
			return
		}
	}

	s.log.Debug("Metrics data received, container id: " + strconv.FormatInt(int64(m.ContainerId), 10))
}
