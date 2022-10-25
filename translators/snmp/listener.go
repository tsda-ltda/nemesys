package snmp

import (
	"context"
	"fmt"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
)

// getListener listen to get metric data messages and redirect to
// the dataProducer.
func (s *SNMPService) getListener() {
	// create socket channel
	ch, err := s.amqp.Channel()
	if err != nil {
		s.Log.Panic("fail to open amqp socket channel", logger.ErrField(err))
		return
	}
	defer ch.Close()

	// declare exchange
	err = ch.ExchangeDeclare(
		amqp.ExchangeGetMetricData, // name
		"direct",                   // type
		true,                       // durable
		false,                      // auto-deleted
		false,                      // internal
		false,                      // no-wait
		nil,                        // arguments
	)
	if err != nil {
		s.Log.Panic("fail to declare exchange", logger.ErrField(err))
		return
	}

	// declare queue
	q, err := ch.QueueDeclare(
		amqp.QueueSNMPGetData, // name
		false,                 // durable
		false,                 // delete when unused
		false,                 // exclusive
		false,                 // no-wait
		nil,                   // arguments
	)
	if err != nil {
		s.Log.Panic("fail to declare queue", logger.ErrField(err))
		return
	}

	// bind queue
	err = ch.QueueBind(
		q.Name,                     // queue name
		"snmp",                     // routing key
		amqp.ExchangeGetMetricData, // exchange
		false,                      // no-wait
		nil,                        // arguments
	)
	if err != nil {
		s.Log.Panic("fail to bind queue", logger.ErrField(err))
		return
	}

	// consume get queue
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
		s.Log.Panic("fail to consume queue", logger.ErrField(err))
		return
	}

	for {
		select {
		case d := <-msgs:
			ctx := context.Background()

			// decode message data
			var m models.GetMetricData
			err = amqp.Decode(d.Body, &m)
			if err != nil {
				s.Log.Error("fail to unmarshal amqp message body", logger.ErrField(err))
				continue
			}

			// check if agent connection exists
			c, ok := s.conns[m.ContainerId]
			if !ok {
				err = s.RegisterAgent(ctx, m.ContainerId)
				if err != nil {
					s.Log.Error("fail to register agent", logger.ErrField(err))
					continue
				}
			} else {
				// reset ttl
				c.Reset()
			}

			// get metric oid
			metric, ok := s.metrics[m.MetricId]
			if !ok {
				err = s.RegisterMetric(ctx, m.MetricId, c.TTL)
				if err != nil {
					s.Log.Error("fail to register metric", logger.ErrField(err))
					continue
				}
			} else {
				metric.Reset()
			}

			// get routing key on message header
			rk, ok := d.Headers["routing_key"].(string)
			if !ok {
				s.Log.Error("fail to make routing_key assetion from message header")
				continue
			}

			// add data request
			s.singleDataReq <- models.AMQPCorrelated[SingleDataReq]{
				CorrelationId: d.CorrelationId,
				RoutingKey:    rk,
				Data: SingleDataReq{
					Conn:   c,
					Metric: m,
					OID:    metric.OID,
				},
			}
			s.Log.Debug("new data request for metric " + fmt.Sprint(m.MetricId))
		case <-s.Done:
			return
		}
	}
}
