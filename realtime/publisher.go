package rts

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/rabbitmq/amqp091-go"
)

func (s *RTS) MetricDataPublisher() {
	// open amqp socket channel
	ch, err := s.amqp.Channel()
	if err != nil {
		s.Log.Panic("fail to open amqp socket channel", logger.ErrField(err))
		return
	}
	defer ch.Close()

	// declare exchange
	err = ch.ExchangeDeclare(
		amqp.ExchangeRTSMetricData, // exchange
		"fanout",                   // type
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

	closed, canceled := amqp.OnChannelClose(ch)
	for {
		select {
		// publish data
		case p := <-s.metricDataPublisherCh:
			// publish
			err = ch.PublishWithContext(context.Background(),
				amqp.ExchangeRTSMetricData, // exchange
				"",                         // routing key
				false,                      // mandatory
				false,                      // immediate
				p,                          // publishing
			)
			if err != nil {
				s.Log.Error("fail to publish message", logger.ErrField(err))
			}

			s.Log.Debug("publishing metric data, metric id: <encoded>")
		case err := <-closed:
			s.Log.Warn("metric data publisher channel closed", logger.ErrField(err))
			return
		case r := <-canceled:
			s.Log.Warn("metric data publisher channel canceled, reason: " + r)
			return
		}
	}
}

func (s *RTS) MetricDataRequestPublisher() {
	// open amqp socket channel
	ch, err := s.amqp.Channel()
	if err != nil {
		s.Log.Panic("fail to open socket channel", logger.ErrField(err))
		return
	}

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

	closed, canceled := amqp.OnChannelClose(ch)
	for {
		select {
		case r := <-s.metricDataRequestCh:
			ctx := context.Background()
			s.Log.Debug("publishing metric data request for: " + r.RoutingKey)

			// publish
			err = ch.PublishWithContext(ctx,
				amqp.ExchangeGetMetricData, // exchange
				r.RoutingKey,               // routing key
				false,                      // mandatory
				false,                      // immediate
				amqp091.Publishing{
					Expiration:    "2000",
					Headers:       amqp091.Table{"routing_key": "rts"},
					CorrelationId: r.CorrelationId,
					Body:          r.Info,
				},
			)
			if err != nil {
				s.Log.Error("fail to publish metric data request")
				continue
			}
			s.Log.Debug("publishing metric data request, metric id: <encoded>")
		case err := <-closed:
			s.Log.Warn("metric data request publisher channel closed", logger.ErrField(err))
			return
		case r := <-canceled:
			s.Log.Warn("metric data request publisher channel canceled, reason: " + r)
			return
		}
	}
}

func (s *RTS) MetricsDataRequestPublisher() {
	// open amqp socket channel
	ch, err := s.amqp.Channel()
	if err != nil {
		s.Log.Panic("fail to open socket channel", logger.ErrField(err))
		return
	}

	// declare exchange
	err = ch.ExchangeDeclare(
		amqp.ExchangeGetMetricsData, // name
		"direct",                    // type
		true,                        // durable
		false,                       // auto-deleted
		false,                       // internal
		false,                       // no-wait
		nil,                         // arguments
	)
	if err != nil {
		s.Log.Panic("fail to declare exchange", logger.ErrField(err))
		return
	}

	closed, canceled := amqp.OnChannelClose(ch)
	for {
		select {
		case r := <-s.metricsDataRequestCh:
			ctx := context.Background()
			s.Log.Debug("publishing metrics data request for: " + r.RoutingKey)

			// publish
			err = ch.PublishWithContext(ctx,
				amqp.ExchangeGetMetricsData, // exchange
				r.RoutingKey,                // routing key
				false,                       // mandatory
				false,                       // immediate
				amqp091.Publishing{
					Expiration:    "2000",
					Headers:       amqp091.Table{"routing_key": "rts"},
					CorrelationId: r.CorrelationId,
					Body:          r.Info,
				},
			)
			if err != nil {
				s.Log.Error("fail to publish metrics data request")
				continue
			}
			s.Log.Debug("publishing metrics data request, container id: <encoded>")
		case err := <-closed:
			s.Log.Warn("metrics data request publisher channel closed", logger.ErrField(err))
			return
		case r := <-canceled:
			s.Log.Warn("metrics data request publisher channel canceled, reason: " + r)
			return
		}
	}
}
