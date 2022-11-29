package amqph

import (
	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/types"
	"github.com/rabbitmq/amqp091-go"
)

type MetricNotification struct {
	// ContainerType is the container type.
	ContainerType types.ContainerType
	// Base is the metric base.
	Base models.BaseMetric
	// Protocol is the metric protocol configuration.
	Protocol any
}

// NotifyMetricUpdated notifies that a metric was updated.
func (a *Amqph) NotifyMetricUpdated(baseMetric models.BaseMetric, protocol any, containerType types.ContainerType) error {
	b, err := amqp.Encode(MetricNotification{
		ContainerType: containerType,
		Base:          baseMetric,
		Protocol:      protocol,
	})
	if err != nil {
		return err
	}

	a.PublisherCh <- models.DetailedPublishing{
		Exchange: amqp.ExchangeMetricUpdated,
		Publishing: amqp091.Publishing{
			Expiration: amqp.DefaultExp,
			Body:       b,
		},
	}
	return nil
}

// NotifyMetricCreated notifies that a metric was created.
func (a *Amqph) NotifyMetricCreated(baseMetric models.BaseMetric, protocol any, containerType types.ContainerType) error {
	b, err := amqp.Encode(MetricNotification{
		ContainerType: containerType,
		Base:          baseMetric,
		Protocol:      protocol,
	})
	if err != nil {
		return err
	}

	a.PublisherCh <- models.DetailedPublishing{
		Exchange: amqp.ExchangeMetricCreated,
		Publishing: amqp091.Publishing{
			Expiration: amqp.DefaultExp,
			Body:       b,
		},
	}
	return nil
}

// NotifyMetricDeleted notifies that a metric was deleted.
func (a *Amqph) NotifyMetricDeleted(id int64, containerId int32) error {
	b, err := amqp.Encode(id)
	if err != nil {
		return err
	}

	a.PublisherCh <- models.DetailedPublishing{
		Exchange: amqp.ExchangeMetricDeleted,
		Publishing: amqp091.Publishing{
			Expiration: amqp.DefaultExp,
			Body:       b,
		},
	}
	return nil
}

func (a *Amqph) OnMetricCreated(queue ...string) <-chan MetricNotification {
	var q string
	if len(queue) > 0 {
		q = queue[0]
	}
	delivery := make(chan MetricNotification)
	go func() {
		msgs, err := a.Listen(q, amqp.ExchangeMetricCreated)
		if err != nil {
			a.log.Panic("Fail to listen amqp messages", logger.ErrField(err))
			return
		}
		for d := range msgs {
			var n MetricNotification
			err = amqp.Decode(d.Body, &n)
			if err != nil {
				a.log.Error("Fail to decode delivery body", logger.ErrField(err))
				continue
			}
			delivery <- n
		}
	}()
	return delivery
}

func (a *Amqph) OnMetricUpdated(queue ...string) <-chan MetricNotification {
	var q string
	if len(queue) > 0 {
		q = queue[0]
	}
	delivery := make(chan MetricNotification)
	go func() {
		msgs, err := a.Listen(q, amqp.ExchangeMetricUpdated)
		if err != nil {
			a.log.Panic("Fail to listen amqp messages", logger.ErrField(err))
			return
		}
		for d := range msgs {
			var n MetricNotification
			err = amqp.Decode(d.Body, &n)
			if err != nil {
				a.log.Error("Fail to decode delivery body", logger.ErrField(err))
				continue
			}
			delivery <- n
		}
	}()
	return delivery
}

func (a *Amqph) OnMetricDeleted() <-chan models.MetricPairId {
	delivery := make(chan models.MetricPairId)
	go func() {
		msgs, err := a.Listen("", amqp.ExchangeMetricDeleted)
		if err != nil {
			a.log.Panic("Fail to listen amqp messages", logger.ErrField(err))
			return
		}
		for d := range msgs {
			var p models.MetricPairId
			err = amqp.Decode(d.Body, &p)
			if err != nil {
				a.log.Error("Fail to decode delivery body", logger.ErrField(err))
				continue
			}
			delivery <- p
		}
	}()
	return delivery
}
