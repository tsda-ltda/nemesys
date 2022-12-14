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

func (s *RTS) onDataPolicyDeleted(id int16) {
	for _, cp := range s.pulling {
		cp.Close()
	}
}

func (s *RTS) onContainerUpdated(base models.BaseContainer, protocol any) {
	c, ok := s.pulling[base.Id]
	if !ok {
		return
	}
	c.Close()
}

func (s *RTS) onContainerDeleted(id int32) {
	c, ok := s.pulling[id]
	if !ok {
		return
	}
	c.Close()
}

func (s *RTS) onMetricUpdated(base models.BaseMetric, protocol any) {
	c, ok := s.pulling[base.ContainerId]
	if !ok {
		return
	}
	m, ok := c.Metrics[base.Id]
	if !ok {
		return
	}
	m.Stop()
}

func (s *RTS) onMetricDeleted(containerId int32, id int64) {
	cp, ok := s.pulling[containerId]
	if !ok {
		return
	}
	m, ok := cp.Metrics[id]
	if !ok {
		return
	}
	m.Stop()
}

func (s *RTS) metricDataRequestListener() {
	var options amqph.ListenerOptions
	options.QueueDeclarationOptions.Name = amqp.QueueRTSMetricDataReq
	options.QueueBindOptions.Exchange = amqp.ExchangeMetricDataReq
	options.QueueBindOptions.RoutingKey = "rts"

	msgs, done := s.amqph.Listen(options)
	for {
		select {
		case d := <-msgs:
			ctx := context.Background()

			if len(d.CorrelationId) == 0 {
				s.log.Warn("Received a rts data request but no correlation id was provided")
				continue
			}

			var r models.MetricRequest
			err := amqp.Decode(d.Body, &r)
			if err != nil {
				s.log.Error("Fail to decode message body", logger.ErrField(err))
				continue
			}

			responseRk, err := amqp.GetRoutingKeyFromHeader(d.Headers)
			if err != nil {
				s.log.Error("Fail to get response routing key from header", logger.ErrField(err))
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
					}, responseRk)
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
				continue
			}

			s.restartMetricPulling(r)
			s.publishRTSMetricData(amqp091.Publishing{
				Expiration:    amqp.DefaultExp,
				Body:          bytes,
				Type:          amqp.FromMessageType(amqp.OK),
				CorrelationId: d.CorrelationId,
			}, responseRk)
			s.log.Debug("Metric data fetched on cache, metric id: " + metricIdString)
		case <-done:
			return
		}
	}
}

func (s *RTS) metricDataListenerHandler(d amqp091.Delivery) {
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

// metricDataListen listen to metric data response, using a unique routing key,
// to resolve rts data requests.
func (s *RTS) metricDataListener() {
	var options amqph.ListenerOptions
	options.QueueDeclarationOptions.Exclusive = true
	options.QueueBindOptions.Exchange = amqp.ExchangeMetricDataRes
	options.QueueBindOptions.RoutingKey = s.GetServiceIdent()

	msgs, done := s.amqph.Listen(options)
	for {
		select {
		case d := <-msgs:
			s.metricDataListenerHandler(d)
		case <-done:
			return
		}
	}
}

// globalMetricDataListener listen to metric data response, using rts routing key,
// to resolve incoming data that are not rts requests responses.
func (s *RTS) globalMetricDataListener() {
	var options amqph.ListenerOptions
	options.QueueDeclarationOptions.Name = amqp.QueueRTSMetricData
	options.QueueBindOptions.Exchange = amqp.ExchangeMetricDataRes
	options.QueueBindOptions.RoutingKey = "rts"

	msgs, done := s.amqph.Listen(options)
	for {
		select {
		case d := <-msgs:
			s.metricDataListenerHandler(d)
		case <-done:
			return
		}
	}
}

// metricsDataListener listen to metrics data response.
func (s *RTS) metricsDataListener() {
	var options amqph.ListenerOptions
	options.QueueDeclarationOptions.Exclusive = true
	options.QueueBindOptions.Exchange = amqp.ExchangeMetricsDataRes
	options.QueueBindOptions.RoutingKey = s.GetServiceIdent()

	msgs, done := s.amqph.Listen(options)
	for {
		select {
		case d := <-msgs:
			if amqp.ToMessageType(d.Type) != amqp.OK {
				continue
			}
			ctx := context.Background()

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
		case <-done:
			return
		}
	}
}
