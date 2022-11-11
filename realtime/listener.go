package rts

import (
	"context"
	"strconv"
	"time"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/db"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/go-redis/redis/v8"
	"github.com/rabbitmq/amqp091-go"
)

func (s *RTS) MetricDataRequestListener() {
	// open amqp socket channel
	ch, err := s.amqp.Channel()
	if err != nil {
		s.Log.Panic("fail to open amqp socket channel", logger.ErrField(err))
		return
	}
	defer ch.Close()

	// declare exchange
	err = ch.ExchangeDeclare(
		amqp.ExchangeRTSGetMetricData, // name
		"direct",                      // type
		true,                          // durable
		false,                         // auto-deleted
		false,                         // internal
		false,                         // no-wait
		nil,                           // arguments
	)
	if err != nil {
		s.Log.Panic("fail to declare exchange", logger.ErrField(err))
		return
	}

	// declare queue
	q, err := ch.QueueDeclare(
		amqp.QueueRTSGetMetricData, // name
		false,                      // durable
		false,                      // delete when unused
		false,                      // exclusive
		false,                      // no-wait
		nil,                        // arguments
	)
	if err != nil {
		s.Log.Panic("fail to declare queue", logger.ErrField(err))
		return
	}

	// bind queue
	err = ch.QueueBind(
		q.Name,                        // queue name
		"",                            // routing key
		amqp.ExchangeRTSGetMetricData, // exchange
		false,                         // no-wait
		nil,                           // args
	)
	if err != nil {
		s.Log.Panic("fail to bind queue", logger.ErrField(err))
		return
	}

	// consume messages
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		s.Log.Panic("fail to consume messages", logger.ErrField(err))
	}

	// close and cancel channels
	closed, canceled := amqp.OnChannelClose(ch)
	for {
		select {
		case d := <-msgs:
			ctx := context.Background()

			// check if correlation id empty
			if len(d.CorrelationId) == 0 {
				s.Log.Warn("receive a rts data request but no correlation id was provided")
				continue
			}

			// decode message body
			var r models.MetricRequest
			err = amqp.Decode(d.Body, &r)
			if err != nil {
				s.Log.Error("fail to decode message body", logger.ErrField(err))
				continue
			}

			// parse metric id to string
			metricIdString := strconv.FormatInt(int64(r.ContainerId), 10)
			s.Log.Debug("get metric data request received, container id: " + metricIdString)

			// get data on cache
			bytes, err := s.cache.Get(ctx, db.RDBCacheMetricDataKey(r.MetricId))
			if err != nil {
				if err != redis.Nil {
					s.Log.Error("fail to get metric data on redis", logger.ErrField(err))
					continue
				}

				// get metric container type and RTS configuration
				res, err := s.pgConn.Metrics.GetRTSConfig(ctx, r.MetricId)
				if err != nil {
					s.Log.Error("fail to get metric rts information", logger.ErrField(err))
					continue
				}

				// check if configuration does not exists
				if !res.Exists {
					s.Log.Warn("fail to get metric rts information, metric does not exist")
					continue
				}

				// start pulling
				s.startMetricPulling(r, res.RTSConfig)

				// publish data when available
				go func(correlationId string, r models.MetricRequest) {
					// send data request
					s.metricDataRequestCh <- models.AMQPCorrelated[[]byte]{
						RoutingKey:    amqp.GetDataRoutingKey(r.ContainerType),
						CorrelationId: d.CorrelationId,
						Info:          d.Body,
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
						s.Log.Warn("plumber timeout, no data response was available")
						return
					}

					// publish data
					s.metricDataPublisherCh <- amqp091.Publishing{
						Type:          res.Type,
						Body:          res.Body,
						CorrelationId: correlationId,
					}
				}(d.CorrelationId, r)
				continue
			}
			s.Log.Debug("metric data fetched on cache, metric id: " + metricIdString)

			// restart/start pulling
			s.restartMetricPulling(r)

			// publish data
			s.metricDataPublisherCh <- amqp091.Publishing{
				Expiration:    "5000",
				Body:          bytes,
				Type:          amqp.FromMessageType(amqp.OK),
				CorrelationId: d.CorrelationId,
			}

		case err := <-closed:
			s.Log.Warn("metric data request listener channel closed", logger.ErrField(err))
			return
		case r := <-canceled:
			s.Log.Warn("metric data request listener channel canceled, reason: " + r)
			return
		}
	}
}

func (s *RTS) MetricDataListener() {
	// open amqp socket channel
	ch, err := s.amqp.Channel()
	if err != nil {
		s.Log.Panic("fail to open amqp socket channel", logger.ErrField(err))
		return
	}
	defer ch.Close()

	// declare exchange
	err = ch.ExchangeDeclare(
		amqp.ExchangeMetricData, // name
		"direct",                // type
		true,                    // durable
		false,                   // auto-deleted
		false,                   // internal
		false,                   // no-wait
		nil,                     // arguments
	)
	if err != nil {
		s.Log.Panic("fail to declare exchange", logger.ErrField(err))
		return
	}

	// declare queue
	q, err := ch.QueueDeclare(
		amqp.QueueRTSMetricData, // name
		false,                   // durable
		false,                   // delete when unused
		true,                    // exclusive
		false,                   // no-wait
		nil,                     // arguments
	)
	if err != nil {
		s.Log.Panic("fail to declare queue", logger.ErrField(err))
		return
	}

	// bind queue
	err = ch.QueueBind(
		q.Name,                  // queue name
		"rts",                   // routing key
		amqp.ExchangeMetricData, // exchange
		false,                   // no-wait
		nil,                     // args
	)
	if err != nil {
		s.Log.Panic("fail to bind queue", logger.ErrField(err))
		return
	}

	// consume messages
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		s.Log.Panic("fail to consume messages", logger.ErrField(err))
	}

	closed, canceled := amqp.OnChannelClose(ch)
	for {
		select {
		case d := <-msgs:
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
				s.Log.Error("fail to decode amqp message body", logger.ErrField(err))
				continue
			}

			// save on cache
			err = s.cache.Set(ctx, d.Body, db.RDBCacheMetricDataKey(m.Id), time.Millisecond*time.Duration(info.CacheDuration))
			if err != nil {
				s.Log.Error("fail to save metric data on cache", logger.ErrField(err))
				continue
			}

			s.Log.Debug("metric data received, id: " + strconv.FormatInt(m.Id, 10))
		case err := <-closed:
			s.Log.Warn("metric data listener channel closed", logger.ErrField(err))
			return
		case r := <-canceled:
			s.Log.Warn("metric data listener channel canceled, reason: " + r)
			return
		}
	}
}

func (s *RTS) MetricsDataListener() {
	// open amqp socket channel
	ch, err := s.amqp.Channel()
	if err != nil {
		s.Log.Panic("fail to open amqp socket channel", logger.ErrField(err))
		return
	}
	defer ch.Close()

	// declare exchange
	err = ch.ExchangeDeclare(
		amqp.ExchangeMetricsData, // name
		"direct",                 // type
		true,                     // durable
		false,                    // auto-deleted
		false,                    // internal
		false,                    // no-wait
		nil,                      // arguments
	)
	if err != nil {
		s.Log.Panic("fail to declare exchange", logger.ErrField(err))
		return
	}

	// declare queue
	q, err := ch.QueueDeclare(
		amqp.QueueRTSMetricsData, // name
		false,                    // durable
		false,                    // delete when unused
		true,                     // exclusive
		false,                    // no-wait
		nil,                      // arguments
	)
	if err != nil {
		s.Log.Panic("fail to declare queue", logger.ErrField(err))
		return
	}

	// bind queue
	err = ch.QueueBind(
		q.Name,                   // queue name
		"rts",                    // routing key
		amqp.ExchangeMetricsData, // exchange
		false,                    // no-wait
		nil,                      // args
	)
	if err != nil {
		s.Log.Panic("fail to bind queue", logger.ErrField(err))
		return
	}

	// consume messages
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		s.Log.Panic("fail to consume messages", logger.ErrField(err))
	}

	closed, canceled := amqp.OnChannelClose(ch)
	for {
		select {
		case d := <-msgs:
			ctx := context.Background()

			// check if message type
			if amqp.ToMessageType(d.Type) != amqp.OK {
				continue
			}

			// decode message body
			var m models.MetricsDataResponse
			err := amqp.Decode(d.Body, &m)
			if err != nil {
				s.Log.Error("fail to decode amqp message body", logger.ErrField(err))
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
					s.Log.Error("fail to encode metric response", logger.ErrField(err))
					continue
				}

				// save on cache
				err = s.cache.Set(ctx, b, db.RDBCacheMetricDataKey(v.Id), time.Millisecond*time.Duration(mp.CacheDuration))
				if err != nil {
					s.Log.Error("fail to save metric data on cache", logger.ErrField(err))
					continue
				}
			}

			s.Log.Debug("metrics data received, container id: " + strconv.FormatInt(int64(m.ContainerId), 10))
		case err := <-closed:
			s.Log.Warn("metric data listener channel closed", logger.ErrField(err))
			return
		case r := <-canceled:
			s.Log.Warn("metric data listener channel canceled, reason: " + r)
			return
		}
	}
}
