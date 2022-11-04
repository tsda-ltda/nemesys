package rts

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/db"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/go-redis/redis/v8"
	"github.com/rabbitmq/amqp091-go"
)

func (s *RTS) DataRequestListener() {
	// open amqp socket channel
	ch, err := s.amqp.Channel()
	if err != nil {
		s.Log.Panic("fail to open amqp socket channel", logger.ErrField(err))
		return
	}
	defer ch.Close()

	// declare exchange
	err = ch.ExchangeDeclare(
		amqp.ExchangeRTSGetData, // name
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
		amqp.QueueRTSGetData, // name
		false,                // durable
		false,                // delete when unused
		false,                // exclusive
		false,                // no-wait
		nil,                  // arguments
	)
	if err != nil {
		s.Log.Panic("fail to declare queue", logger.ErrField(err))
		return
	}

	// bind queue
	err = ch.QueueBind(
		q.Name,                  // queue name
		"",                      // routing key
		amqp.ExchangeRTSGetData, // exchange
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
			s.Log.Debug("new data request, metric id: " + strconv.FormatInt(r.MetricId, 10))

			// get data on cache
			bytes, err := s.cache.Get(ctx, db.RDBCacheMetricDataKey(r.MetricId))
			if err != nil {
				if err != redis.Nil {
					s.Log.Error("fail to get metric data on redis", logger.ErrField(err))
					continue
				}

				// get metric container type and RTS configuration
				e, info, err := s.pgConn.Metrics.GetRTSConfig(ctx, r.MetricId)
				if err != nil {
					s.Log.Error("fail to get metric rts information", logger.ErrField(err))
					continue
				}

				// check if configuration does not exists
				if !e {
					s.Log.Warn("fail to get metric rts information, metric does not exist")
					continue
				}

				// do pulling stuff here...

				// publish data when available
				go func(correlationId string, r models.MetricRequest) {
					// send data request
					s.getDataCh <- models.AMQPCorrelated[[]byte]{
						RoutingKey:    amqp.GetDataRoutingKey(r.ContainerType),
						CorrelationId: d.CorrelationId,
						Info:          d.Body,
					}

					// set pending request
					s.pendingDataMap[correlationId] = info

					// delete channel
					defer func(correlationId string) {
						delete(s.pendingDataMap, correlationId)
					}(correlationId)

					// wait response
					res, err := s.plumber.Listen(correlationId, time.Second*5)
					if err != nil {
						s.Log.Warn("plumber timeouted, no data response was available")
						return
					}

					// publish data
					s.publisherDataCh <- amqp091.Publishing{
						Type:          res.Type,
						Body:          res.Body,
						CorrelationId: correlationId,
					}
				}(d.CorrelationId, r)
				continue
			}

			// publish data
			s.publisherDataCh <- amqp091.Publishing{
				Body:          bytes,
				Type:          amqp.FromMessageType(amqp.OK),
				CorrelationId: d.CorrelationId,
			}

		case <-s.stopDataRequestListener:
			return
		}
	}
}

func (s *RTS) DataListener() {
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
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
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

	for {
		select {
		case d := <-msgs:
			ctx := context.Background()

			// check for pending request
			info, ok := s.pendingDataMap[d.CorrelationId]

			// send data as response for a possible listener
			s.plumber.Send(d)
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
			s.Log.Debug("new data received, metric id: " + strconv.FormatInt(m.MetricId, 10))

			// check if info exists
			if !ok {
				continue
			}

			// save on cache
			err = s.cache.Set(ctx, d.Body, db.RDBCacheMetricDataKey(m.MetricId), time.Millisecond*time.Duration(info.CacheDuration))
			if err != nil {
				s.Log.Error("fail to save metric data on cache", logger.ErrField(err))
				continue
			}
			s.Log.Debug("metric data saved on cache, metricId: " + fmt.Sprint(m.MetricId))
		case <-s.stopDataListener:
			return
		}
	}
}
