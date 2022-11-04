package rts

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/rabbitmq/amqp091-go"
)

func (s *RTS) DataPublisher() {
	// open amqp socket channel
	ch, err := s.amqp.Channel()
	if err != nil {
		s.Log.Panic("fail to open amqp socket channel", logger.ErrField(err))
		return
	}
	defer ch.Close()

	// declare exchange
	err = ch.ExchangeDeclare(
		amqp.ExchangeRTSData, // exchange
		"fanout",             // type
		true,                 // durable
		false,                // auto-deleted
		false,                // internal
		false,                // no-wait
		nil,                  // arguments
	)
	if err != nil {
		s.Log.Panic("fail to declare exchange", logger.ErrField(err))
		return
	}

	for {
		select {
		// publish data
		case p := <-s.publisherDataCh:
			s.Log.Debug("publishing new data for listeners")

			// publish
			err = ch.PublishWithContext(context.Background(),
				amqp.ExchangeRTSData, // exchange
				"",                   // routing key
				false,                // mandatory
				false,                // immediate
				p,                    // publishing
			)
			if err != nil {
				s.Log.Error("fail to publish message", logger.ErrField(err))
			}
		// quit
		case <-s.stopDataPublisher:
			return
		}
	}
}

func (s *RTS) DataRequestPublisher() {
	// open amqp socket channel
	ch, err := s.amqp.Channel()
	if err != nil {
		s.Log.Panic("fail to open socket channel", logger.ErrField(err))
		return
	}

	// declare snmp exchange
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

	for {
		select {
		// get data
		case r := <-s.getDataCh:
			ctx := context.Background()
			s.Log.Debug("publishing new data request for translators, routing key: " + r.RoutingKey)

			// publish
			ch.PublishWithContext(ctx,
				amqp.ExchangeGetMetricData, // exchange
				r.RoutingKey,               // routing key
				false,                      // mandatory
				false,                      // immediate
				amqp091.Publishing{
					Headers:       amqp091.Table{"routing_key": "rts"},
					CorrelationId: r.CorrelationId,
					Body:          r.Info,
				},
			)
		case <-s.stopDataRequestPublisher:
			return
		}
	}
}
