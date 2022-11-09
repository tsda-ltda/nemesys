package amqph

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/rabbitmq/amqp091-go"
)

// metricNotifier receives updates through metricNotifierCh and sends a fanout amqp message to notify that a metric
// has been created or updated.
func (a *Amqph) metricNotifier() {
	// open socket channel
	ch, err := a.conn.Channel()
	if err != nil {
		a.log.Panic("fail to open socket channel", logger.ErrField(err))
	}

	// declare get data exchange
	err = ch.ExchangeDeclare(
		amqp.ExchangeNotifyMetric, // name
		"fanout",                  // type
		true,                      // durable
		false,                     // auto-deleted
		false,                     // internal
		false,                     // no-wait
		nil,                       // arguments
	)
	if err != nil {
		a.log.Panic("fail to declare exchange", logger.ErrField(err))
	}

	// close and cancel channels
	closedCh := make(chan *amqp091.Error)
	canceledCh := make(chan string)
	ch.NotifyCancel(canceledCh)
	ch.NotifyClose(closedCh)

	for {
		select {
		case n := <-a.metricNotifierCh:
			// encode data
			b, err := amqp.Encode(n)
			if err != nil {
				a.log.Error("fail to encode metric notification data", logger.ErrField(err))
				continue
			}

			// publish message
			err = ch.PublishWithContext(context.Background(),
				amqp.ExchangeRTSGetMetricData, // exchange
				"",                            // routing key
				false,                         // mandatory
				false,                         // immediate
				amqp091.Publishing{
					Body: b,
				},
			)
			if err != nil {
				a.log.Error("fail to publish metric notification", logger.ErrField(err))
				continue
			}
			a.log.Debug("metric notification sent with success")

		case err := <-closedCh:
			a.log.Warn("metric notification channel closed", logger.ErrField(err))
			return
		case r := <-canceledCh:
			a.log.Warn("metric notification channel canceled, reason: " + r)
			return
		}
	}
}

// NotifyMetric notifies that a metric have been created or updated.
func (a *Amqph) NotifyMetric(metric any) {
	a.metricNotifierCh <- metric
}
