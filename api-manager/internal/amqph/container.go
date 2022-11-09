package amqph

import (
	"context"
	"strconv"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/types"
	"github.com/rabbitmq/amqp091-go"
)

type containerNotification struct {
	// Type is the container type.
	Type types.ContainerType
	// Data is the container data.
	Data any
}

// containerNotifier receives updates through containerNotifierCH and sends a fanout amqp message to notify that a container
// has been created or updated.
func (a *Amqph) containerNotifier() {
	// open socket channel
	ch, err := a.conn.Channel()
	if err != nil {
		a.log.Panic("fail to open socket channel", logger.ErrField(err))
	}

	// declare get data exchange
	err = ch.ExchangeDeclare(
		amqp.ExchangeNotifyContainer, // name
		"fanout",                     // type
		true,                         // durable
		false,                        // auto-deleted
		false,                        // internal
		false,                        // no-wait
		nil,                          // arguments
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
		case n := <-a.containerNotifierCh:
			// encode data
			b, err := amqp.Encode(n.Data)
			if err != nil {
				a.log.Error("fail to encode container notification data", logger.ErrField(err))
				continue
			}

			// publish message
			err = ch.PublishWithContext(context.Background(),
				amqp.ExchangeRTSGetMetricData, // exchange
				"",                            // routing key
				false,                         // mandatory
				false,                         // immediate
				amqp091.Publishing{
					Type: strconv.FormatInt(int64(n.Type), 10),
					Body: b,
				},
			)
			if err != nil {
				a.log.Error("fail to publish container notification", logger.ErrField(err))
				continue
			}
			a.log.Debug("container notification sent with success")

		case err := <-closedCh:
			a.log.Warn("container notification channel closed", logger.ErrField(err))
			return
		case r := <-canceledCh:
			a.log.Warn("container notification channel canceled, reason: " + r)
			return
		}
	}
}

// NotifyContainer notifies that a container have been created or updated.
func (a *Amqph) NotifyContainer(container any, containerType types.ContainerType) {
	a.containerNotifierCh <- containerNotification{
		Type: containerType,
		Data: container,
	}
}
