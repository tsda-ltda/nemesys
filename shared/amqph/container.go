package amqph

import (
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

// NotifyContainerUpdated notifies that a container was updated.
func (a *Amqph) NotifyContainerUpdated(baseContainer models.BaseContainer, protocol any, containerType types.ContainerType) error {
	b, err := amqp.Encode(ContainerNotification{
		Type:     containerType,
		Base:     baseContainer,
		Protocol: protocol,
	})
	if err != nil {
		return err
	}

	a.PublisherCh <- models.DetailedPublishing{
		Exchange: amqp.ExchangeContainerUpdated,
		Publishing: amqp091.Publishing{
			Expiration: amqp.DefaultExp,
			Body:       b,
		},
	}
	return nil
}

// NotifyContainerCreated notifies that a container was created.
func (a *Amqph) NotifyContainerCreated(baseContainer models.BaseContainer, protocol any, containerType types.ContainerType) error {
	b, err := amqp.Encode(ContainerNotification{
		Type:     containerType,
		Base:     baseContainer,
		Protocol: protocol,
	})
	if err != nil {
		return err
	}

	a.PublisherCh <- models.DetailedPublishing{
		Exchange: amqp.ExchangeContainerUpdated,
		Publishing: amqp091.Publishing{
			Expiration: amqp.DefaultExp,
			Body:       b,
		},
	}
	return nil
}

// NotifyContainerDeleted notifies that a container was deleted.
func (a *Amqph) NotifyContainerDeleted(id int32) error {
	b, err := amqp.Encode(id)
	if err != nil {
		return err
	}

	a.PublisherCh <- models.DetailedPublishing{
		Exchange: amqp.ExchangeContainerDeleted,
		Publishing: amqp091.Publishing{
			Expiration: amqp.DefaultExp,
			Body:       b,
		},
	}
	return nil
}

func (a *Amqph) OnContainerCreated() <-chan ContainerNotification {
	delivery := make(chan ContainerNotification)
	go func() {
		msgs, err := a.Listen("", amqp.ExchangeContainerCreated)
		if err != nil {
			a.log.Panic("fail to listen amqp messages", logger.ErrField(err))
			return
		}
		for d := range msgs {
			var n ContainerNotification
			err = amqp.Decode(d.Body, &n)
			if err != nil {
				a.log.Error("fail to decode delivery body", logger.ErrField(err))
				continue
			}
			delivery <- n
		}
	}()
	return delivery
}

func (a *Amqph) OnContainerUpdated() <-chan ContainerNotification {
	delivery := make(chan ContainerNotification)
	go func() {
		msgs, err := a.Listen("", amqp.ExchangeContainerUpdated)
		if err != nil {
			a.log.Panic("fail to listen amqp messages", logger.ErrField(err))
			return
		}
		for d := range msgs {
			var n ContainerNotification
			err = amqp.Decode(d.Body, &n)
			if err != nil {
				a.log.Error("fail to decode delivery body", logger.ErrField(err))
				continue
			}
			delivery <- n
		}
	}()
	return delivery
}

func (a *Amqph) OnContainerDeleted() <-chan int32 {
	delivery := make(chan int32)
	go func() {
		msgs, err := a.Listen("", amqp.ExchangeContainerDeleted)
		if err != nil {
			a.log.Panic("fail to listen amqp messages", logger.ErrField(err))
			return
		}
		for d := range msgs {
			var id int32
			err = amqp.Decode(d.Body, &id)
			if err != nil {
				a.log.Error("fail to decode delivery body", logger.ErrField(err))
				continue
			}
			delivery <- id
		}
	}()
	return delivery
}
