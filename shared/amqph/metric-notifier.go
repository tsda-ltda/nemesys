package amqph

import (
	"context"

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

// MetricNotifier receives notifications and sends a fanout amqp message to notify others services.
func (a *Amqph) MetricNotifier() {
	// open socket channel
	ch, err := a.conn.Channel()
	if err != nil {
		a.log.Panic("fail to open socket channel", logger.ErrField(err))
	}

	// declare exchanges
	err = ch.ExchangeDeclare(
		amqp.ExchangeNotifyCreatedMetric, // name
		"fanout",                         // type
		true,                             // durable
		false,                            // auto-deleted
		false,                            // internal
		false,                            // no-wait
		nil,                              // arguments
	)
	if err != nil {
		a.log.Panic("fail to declare exchange", logger.ErrField(err))
	}

	err = ch.ExchangeDeclare(
		amqp.ExchangeNotifyUpdatedMetric, // name
		"fanout",                         // type
		true,                             // durable
		false,                            // auto-deleted
		false,                            // internal
		false,                            // no-wait
		nil,                              // arguments
	)
	if err != nil {
		a.log.Panic("fail to declare exchange", logger.ErrField(err))
	}

	err = ch.ExchangeDeclare(
		amqp.ExchangeNotifyDeletedMetric, // name
		"fanout",                         // type
		true,                             // durable
		false,                            // auto-deleted
		false,                            // internal
		false,                            // no-wait
		nil,                              // arguments
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
		case n := <-a.metricCreatedNotifierCh:
			// encode data
			b, err := amqp.Encode(n)
			if err != nil {
				a.log.Error("fail to encode metric notification data", logger.ErrField(err))
				continue
			}

			// publish message
			err = ch.PublishWithContext(context.Background(),
				amqp.ExchangeNotifyCreatedMetric, // exchange
				"",                               // routing key
				false,                            // mandatory
				false,                            // immediate
				amqp091.Publishing{
					Body: b,
				},
			)
			if err != nil {
				a.log.Error("fail to publish metric creation notification", logger.ErrField(err))
				continue
			}
			a.log.Debug("notification of a metric creation was published")
		case n := <-a.metricUpdatedNotifierCh:
			// encode data
			b, err := amqp.Encode(n)
			if err != nil {
				a.log.Error("fail to encode metric notification data", logger.ErrField(err))
				continue
			}

			// publish message
			err = ch.PublishWithContext(context.Background(),
				amqp.ExchangeNotifyUpdatedMetric, // exchange
				"",                               // routing key
				false,                            // mandatory
				false,                            // immediate
				amqp091.Publishing{
					Body: b,
				},
			)
			if err != nil {
				a.log.Error("fail to publish metric update notification", logger.ErrField(err))
				continue
			}
			a.log.Debug("notification of a metric update was published")
		case mp := <-a.metricDeletedNotifierCh:
			// encode data
			b, err := amqp.Encode(mp)
			if err != nil {
				a.log.Error("fail to encode metric notification data", logger.ErrField(err))
				continue
			}

			// publish message
			err = ch.PublishWithContext(context.Background(),
				amqp.ExchangeNotifyDeletedMetric, // exchange
				"",                               // routing key
				false,                            // mandatory
				false,                            // immediate
				amqp091.Publishing{
					Body: b,
				},
			)
			if err != nil {
				a.log.Error("fail to publish metric deleted notification", logger.ErrField(err))
				continue
			}
			a.log.Debug("notification of a metric deleted was published")
		case err := <-closedCh:
			a.log.Warn("metric notification channel closed", logger.ErrField(err))
			return
		case r := <-canceledCh:
			a.log.Warn("metric notification channel canceled, reason: " + r)
			return
		}
	}
}

// NotifyMetricUpdated notifies that a metric was updated.
func (a *Amqph) NotifyMetricUpdated(baseMetric models.BaseMetric, protocol any, containerType types.ContainerType) {
	a.metricUpdatedNotifierCh <- MetricNotification{
		ContainerType: containerType,
		Base:          baseMetric,
		Protocol:      protocol,
	}
}

// NotifyMetricCreated notifies that a metric was created.
func (a *Amqph) NotifyMetricCreated(baseMetric models.BaseMetric, protocol any, containerType types.ContainerType) {
	a.metricUpdatedNotifierCh <- MetricNotification{
		ContainerType: containerType,
		Base:          baseMetric,
		Protocol:      protocol,
	}
}

// NotifyMetricDeleted notifies that a metric was deleted.
func (a *Amqph) NotifyMetricDeleted(id int64, containerId int32) {
	a.metricDeletedNotifierCh <- models.MetricPairId{
		Id:          id,
		ContainerId: containerId,
	}
}
