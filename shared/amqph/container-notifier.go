package amqph

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/types"
	"github.com/rabbitmq/amqp091-go"
)

type ContainerNotification struct {
	// Type is the container type.
	Type types.ContainerType
	// Base is the container base.
	Base models.BaseContainer
	// Protocol is the container protocol configuration.
	Protocol any
}

// ContainerNotifier receives notifications and sends a fanout amqp message to notify other services.
func (a *Amqph) ContainerNotifier() {
	// open socket channel
	ch, err := a.conn.Channel()
	if err != nil {
		a.log.Panic("fail to open socket channel", logger.ErrField(err))
	}

	// declare exchanges
	err = ch.ExchangeDeclare(
		amqp.ExchangeNotifyCreatedContainer, // name
		"fanout",                            // type
		true,                                // durable
		false,                               // auto-deleted
		false,                               // internal
		false,                               // no-wait
		nil,                                 // arguments
	)
	if err != nil {
		a.log.Panic("fail to declare exchange", logger.ErrField(err))
	}

	err = ch.ExchangeDeclare(
		amqp.ExchangeNotifyUpdatedContainer, // name
		"fanout",                            // type
		true,                                // durable
		false,                               // auto-deleted
		false,                               // internal
		false,                               // no-wait
		nil,                                 // arguments
	)
	if err != nil {
		a.log.Panic("fail to declare exchange", logger.ErrField(err))
	}

	err = ch.ExchangeDeclare(
		amqp.ExchangeNotifyDeletedContainer, // name
		"fanout",                            // type
		true,                                // durable
		false,                               // auto-deleted
		false,                               // internal
		false,                               // no-wait
		nil,                                 // arguments
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
		case n := <-a.containerCreatedNotifierCh:
			// encode data
			b, err := amqp.Encode(n)
			if err != nil {
				a.log.Error("fail to encode container notification data", logger.ErrField(err))
				continue
			}

			// publish message
			err = ch.PublishWithContext(context.Background(),
				amqp.ExchangeNotifyCreatedContainer, // exchange
				"",                                  // routing key
				false,                               // mandatory
				false,                               // immediate
				amqp091.Publishing{
					Body: b,
				},
			)
			if err != nil {
				a.log.Error("fail to publish container creation notification", logger.ErrField(err))
				continue
			}
			a.log.Debug("notification of a container creation was published")
		case n := <-a.containerUpdatesNotifierCh:
			// encode data
			b, err := amqp.Encode(n)
			if err != nil {
				a.log.Error("fail to encode container notification data", logger.ErrField(err))
				continue
			}

			// publish message
			err = ch.PublishWithContext(context.Background(),
				amqp.ExchangeNotifyUpdatedContainer, // exchange
				"",                                  // routing key
				false,                               // mandatory
				false,                               // immediate
				amqp091.Publishing{
					Body: b,
				},
			)
			if err != nil {
				a.log.Error("fail to publish container update notification", logger.ErrField(err))
				continue
			}
			a.log.Debug("notification of a container update was published")
		case id := <-a.containerDeletedNotifierCh:
			// encode data
			b, err := amqp.Encode(id)
			if err != nil {
				a.log.Error("fail to encode container notification data", logger.ErrField(err))
				continue
			}

			// publish message
			err = ch.PublishWithContext(context.Background(),
				amqp.ExchangeNotifyDeletedContainer, // exchange
				"",                                  // routing key
				false,                               // mandatory
				false,                               // immediate
				amqp091.Publishing{
					Body: b,
				},
			)
			if err != nil {
				a.log.Error("fail to publish container deleted notification", logger.ErrField(err))
				continue
			}
			a.log.Debug("notification of a container deleted was published")
		case err := <-closedCh:
			a.log.Warn("container notification channel closed", logger.ErrField(err))
			return
		case r := <-canceledCh:
			a.log.Warn("container notification channel canceled, reason: " + r)
			return
		}
	}
}

// NotifyContainerUpdated notifies that a container was updated.
func (a *Amqph) NotifyContainerUpdated(baseContainer models.BaseContainer, protocol any, containerType types.ContainerType) {
	a.containerUpdatesNotifierCh <- ContainerNotification{
		Type:     containerType,
		Base:     baseContainer,
		Protocol: protocol,
	}
}

// NotifyContainerCreated notifies that a container was created.
func (a *Amqph) NotifyContainerCreated(baseContainer models.BaseContainer, protocol any, containerType types.ContainerType) {
	a.containerCreatedNotifierCh <- ContainerNotification{
		Type:     containerType,
		Base:     baseContainer,
		Protocol: protocol,
	}
}

// NotifyContainerDeleted notifies that a container was deleted.
func (a *Amqph) NotifyContainerDeleted(id int32) {
	a.containerDeletedNotifierCh <- id
}
