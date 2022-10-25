package snmp

import (
	"context"
	"fmt"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/rabbitmq/amqp091-go"
)

type SingleDataReq struct {
	Conn   *Conn
	Metric models.GetMetricData
	OID    string
}

// dataProducer receives data request through SNMPService.dataReq and fetch data asynchronously,
// publishing it in the amqp server.
func (s *SNMPService) dataProducer() {
	// open socket channel
	ch, err := s.amqp.Channel()
	if err != nil {
		s.Log.Panic("fail to open socket channel", logger.ErrField(err))
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

	for {
		select {
		case req := <-s.singleDataReq:
			go func(ch *amqp091.Channel, req models.AMQPCorrelated[SingleDataReq]) {
				ctx := context.Background()

				// fetch data
				r, err := s.Get(req.Data.Conn, []string{req.Data.OID})
				if err != nil {
					s.Log.Warn("fail to get data", logger.ErrField(err))
					return
				}

				// encode data
				bytes, err := amqp.Encode(r)
				if err != nil {
					s.Log.Error("fail to marshal data response", logger.ErrField(err))
					return
				}

				if err != nil {
					s.Log.Error("fail to marshal amqp message", logger.ErrField(err))
					return
				}

				err = ch.PublishWithContext(ctx,
					amqp.ExchangeMetricData, // exchange
					req.RoutingKey,          // routing key
					false,                   // mandatory
					false,                   // immediate
					amqp091.Publishing{
						ContentType:   "application/msgpack",
						CorrelationId: req.CorrelationId,
						Body:          bytes,
					})
				if err != nil {
					s.Log.Error("fail to publish data", logger.ErrField(err))
					return
				}
				s.Log.Debug("data published for metric: " + fmt.Sprint(req.Data.Metric.MetricId))
			}(ch, req)
		case <-s.Done:
			return
		}
	}
}
