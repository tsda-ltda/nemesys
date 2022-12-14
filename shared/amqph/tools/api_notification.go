package tools

import (
	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/amqph"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/rabbitmq/amqp091-go"
)

type ContainerNotification struct {
	Base     models.BaseContainer
	Protocol any
}

type MetricNotification struct {
	Base     models.BaseMetric
	Protocol any
}

type NotificationHandler struct {
	ContainerCreatedQueue string
	MetricCreatedQueue    string

	OnContainerCreated func(base models.BaseContainer, protocol any)
	OnContainerUpdated func(base models.BaseContainer, protocol any)
	OnContainerDeleted func(int32)

	OnMetricCreated func(base models.BaseMetric, protocol any)
	OnMetricUpdated func(base models.BaseMetric, protocol any)
	OnMetricDeleted func(containerId int32, metricId int64)

	OnDataPolicyDeleted func(int16)

	OnError func(error)
}

func HandleAPINotifications(a *amqph.Amqph, handler *NotificationHandler) {
	if handler.OnContainerCreated != nil {
		go containerCreatedListener(a, handler)
	}
	if handler.OnContainerUpdated != nil {
		go containerUpdatedListener(a, handler)
	}
	if handler.OnContainerDeleted != nil {
		go containerDeletedListener(a, handler)
	}
	if handler.OnMetricCreated != nil {
		go metricCreatedListener(a, handler)
	}
	if handler.OnMetricUpdated != nil {
		go metricUpdatedListener(a, handler)
	}
	if handler.OnMetricDeleted != nil {
		go metricDeletedListener(a, handler)
	}
	if handler.OnDataPolicyDeleted != nil {
		go dataPolicyDeletedListener(a, handler)
	}
}

func containerCreatedListener(a *amqph.Amqph, handler *NotificationHandler) {
	var options amqph.ListenerOptions
	if handler.ContainerCreatedQueue != "" {
		options.QueueDeclarationOptions.Name = handler.ContainerCreatedQueue
	} else {
		options.QueueDeclarationOptions.Exclusive = true
	}
	options.QueueBindOptions.Exchange = amqp.ExchangeContainerCreated

	msgs, done := a.Listen(options)
	for {
		select {
		case d := <-msgs:
			var n ContainerNotification
			err := amqp.Decode(d.Body, &n)
			if err != nil {
				handler.OnError(err)
				continue
			}
			handler.OnContainerCreated(n.Base, n.Protocol)
		case <-done:
			return
		}
	}
}

func containerUpdatedListener(a *amqph.Amqph, handler *NotificationHandler) {
	var options amqph.ListenerOptions

	options.QueueDeclarationOptions.Exclusive = true
	options.QueueBindOptions.Exchange = amqp.ExchangeContainerUpdated

	msgs, done := a.Listen(options)
	for {
		select {
		case d := <-msgs:
			var n ContainerNotification
			err := amqp.Decode(d.Body, &n)
			if err != nil {
				handler.OnError(err)
				continue
			}
			handler.OnContainerUpdated(n.Base, n.Protocol)
		case <-done:
			return
		}
	}
}

func containerDeletedListener(a *amqph.Amqph, handler *NotificationHandler) {
	var options amqph.ListenerOptions

	options.QueueDeclarationOptions.Exclusive = true
	options.QueueBindOptions.Exchange = amqp.ExchangeContainerDeleted

	msgs, done := a.Listen(options)
	for {
		select {
		case d := <-msgs:
			var id int32
			err := amqp.Decode(d.Body, &id)
			if err != nil {
				handler.OnError(err)
				continue
			}
			handler.OnContainerDeleted(id)
		case <-done:
			return
		}
	}
}

func metricCreatedListener(a *amqph.Amqph, handler *NotificationHandler) {
	var options amqph.ListenerOptions
	if handler.ContainerCreatedQueue != "" {
		options.QueueDeclarationOptions.Name = handler.MetricCreatedQueue
	} else {
		options.QueueDeclarationOptions.Exclusive = true
	}
	options.QueueBindOptions.Exchange = amqp.ExchangeMetricCreated

	msgs, done := a.Listen(options)
	for {
		select {
		case d := <-msgs:
			var n MetricNotification
			err := amqp.Decode(d.Body, &n)
			if err != nil {
				handler.OnError(err)
				continue
			}
			handler.OnMetricCreated(n.Base, n.Protocol)
		case <-done:
			return
		}
	}
}

func metricUpdatedListener(a *amqph.Amqph, handler *NotificationHandler) {
	var options amqph.ListenerOptions

	options.QueueDeclarationOptions.Exclusive = true
	options.QueueBindOptions.Exchange = amqp.ExchangeMetricUpdated

	msgs, done := a.Listen(options)
	for {
		select {
		case d := <-msgs:
			var n MetricNotification
			err := amqp.Decode(d.Body, &n)
			if err != nil {
				handler.OnError(err)
				continue
			}
			handler.OnMetricUpdated(n.Base, n.Protocol)
		case <-done:
			return
		}
	}
}

func metricDeletedListener(a *amqph.Amqph, handler *NotificationHandler) {
	var options amqph.ListenerOptions

	options.QueueDeclarationOptions.Exclusive = true
	options.QueueBindOptions.Exchange = amqp.ExchangeMetricDeleted

	msgs, done := a.Listen(options)
	for {
		select {
		case d := <-msgs:
			var pair models.MetricPairId
			err := amqp.Decode(d.Body, &pair)
			if err != nil {
				handler.OnError(err)
				continue
			}
			handler.OnMetricDeleted(pair.ContainerId, pair.Id)
		case <-done:
			return
		}
	}
}

func dataPolicyDeletedListener(a *amqph.Amqph, handler *NotificationHandler) {
	var options amqph.ListenerOptions

	options.QueueDeclarationOptions.Exclusive = true
	options.QueueBindOptions.Exchange = amqp.ExchangeDataPolicyDeleted

	msgs, done := a.Listen(options)
	for {
		select {
		case d := <-msgs:
			var id int16
			err := amqp.Decode(d.Body, &id)
			if err != nil {
				handler.OnError(err)
				continue
			}
			handler.OnDataPolicyDeleted(id)
		case <-done:
			return
		}
	}
}

func NotifyContainerCreated(a *amqph.Amqph, base models.BaseContainer, protocol any) (err error) {
	b, err := amqp.Encode(ContainerNotification{
		Base:     base,
		Protocol: protocol,
	})
	if err != nil {
		return err
	}

	a.Publish(amqph.Publish{
		Exchange: amqp.ExchangeContainerCreated,
		Publishing: amqp091.Publishing{
			Body: b,
		},
	})

	return nil
}

func NotifyContainerUpdated(a *amqph.Amqph, base models.BaseContainer, protocol any) (err error) {
	b, err := amqp.Encode(ContainerNotification{
		Base:     base,
		Protocol: protocol,
	})
	if err != nil {
		return err
	}

	a.Publish(amqph.Publish{
		Exchange: amqp.ExchangeContainerUpdated,
		Publishing: amqp091.Publishing{
			Body: b,
		},
	})

	return nil
}

func NotifyMetricCreated(a *amqph.Amqph, base models.BaseMetric, protocol any) (err error) {
	b, err := amqp.Encode(MetricNotification{
		Base:     base,
		Protocol: protocol,
	})
	if err != nil {
		return err
	}

	a.Publish(amqph.Publish{
		Exchange: amqp.ExchangeMetricCreated,
		Publishing: amqp091.Publishing{
			Body: b,
		},
	})

	return nil
}

func NotifyMetricUpdated(a *amqph.Amqph, base models.BaseMetric, protocol any) (err error) {
	b, err := amqp.Encode(MetricNotification{
		Base:     base,
		Protocol: protocol,
	})
	if err != nil {
		return err
	}

	a.Publish(amqph.Publish{
		Exchange: amqp.ExchangeMetricUpdated,
		Publishing: amqp091.Publishing{
			Body: b,
		},
	})

	return nil
}

func NotifyContainerDeleted(a *amqph.Amqph, id int32) (err error) {
	b, err := amqp.Encode(id)
	if err != nil {
		return err
	}

	a.Publish(amqph.Publish{
		Exchange: amqp.ExchangeContainerDeleted,
		Publishing: amqp091.Publishing{
			Body: b,
		},
	})

	return nil
}

func NotifyMetricDeleted(a *amqph.Amqph, containerId int32, id int64) (err error) {
	b, err := amqp.Encode(models.MetricPairId{
		Id:          id,
		ContainerId: containerId,
	})
	if err != nil {
		return err
	}

	a.Publish(amqph.Publish{
		Exchange: amqp.ExchangeMetricDeleted,
		Publishing: amqp091.Publishing{
			Body: b,
		},
	})

	return nil
}

func NotifyDataPolicyDeleted(a *amqph.Amqph, id int16) (err error) {
	b, err := amqp.Encode(id)
	if err != nil {
		return err
	}

	a.Publish(amqph.Publish{
		Exchange: amqp.ExchangeDataPolicyDeleted,
		Publishing: amqp091.Publishing{
			Body: b,
		},
	})

	return nil
}
