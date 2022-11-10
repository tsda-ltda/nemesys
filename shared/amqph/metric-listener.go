package amqph

import (
	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/rabbitmq/amqp091-go"
)

// MetricListener receives notifications and calls a callback.
func (a *Amqph) MetricListener() {
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

	// create queues
	qCreate, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		a.log.Panic("fail to declare queue")
	}

	qUpdate, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		a.log.Panic("fail to declare queue")
	}

	qDelete, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		a.log.Panic("fail to declare queue")
	}

	// bind queues
	err = ch.QueueBind(
		qCreate.Name,                     // queue name
		"",                               // routing key
		amqp.ExchangeNotifyCreatedMetric, // exchange
		false,
		nil,
	)
	if err != nil {
		a.log.Panic("fail to bind queue")
	}

	err = ch.QueueBind(
		qUpdate.Name,                     // queue name
		"",                               // routing key
		amqp.ExchangeNotifyUpdatedMetric, // exchange
		false,
		nil,
	)
	if err != nil {
		a.log.Panic("fail to bind queue")
	}

	err = ch.QueueBind(
		qDelete.Name,                     // queue name
		"",                               // routing key
		amqp.ExchangeNotifyDeletedMetric, // exchange
		false,
		nil,
	)
	if err != nil {
		a.log.Panic("fail to bind queue")
	}

	// consume queues
	msgsCreate, err := ch.Consume(
		qCreate.Name, // queue
		"",           // consumer
		true,         // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	if err != nil {
		a.log.Panic("fail to consume queue")
	}

	msgsUpdate, err := ch.Consume(
		qUpdate.Name, // queue
		"",           // consumer
		true,         // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	if err != nil {
		a.log.Panic("fail to consume queue")
	}

	msgsDelete, err := ch.Consume(
		qDelete.Name, // queue
		"",           // consumer
		true,         // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	if err != nil {
		a.log.Panic("fail to consume queue")
	}

	// close and cancel channels
	closedCh := make(chan *amqp091.Error)
	canceledCh := make(chan string)
	ch.NotifyCancel(canceledCh)
	ch.NotifyClose(closedCh)

	for {
		select {
		case d := <-msgsCreate:
			// decode message
			var n MetricNotification
			err = amqp.Decode(d.Body, &n)
			if err != nil {
				a.log.Error("fail to decode metric notification message body", logger.ErrField(err))
				continue
			}
			a.Notifications.Publish(MetricCreated, n)
		case d := <-msgsUpdate:
			var n MetricNotification
			err = amqp.Decode(d.Body, &n)
			if err != nil {
				a.log.Error("fail to decode metric notification message body", logger.ErrField(err))
				continue
			}
			a.Notifications.Publish(MetricUpdated, n)
		case d := <-msgsDelete:
			var mp models.MetricPairId
			err = amqp.Decode(d.Body, &mp)
			if err != nil {
				a.log.Error("fail to decode metric notification message body", logger.ErrField(err))
				continue
			}
			a.Notifications.Publish(MetricDeleted, mp)
		case err := <-closedCh:
			a.log.Warn("container listener channel closed", logger.ErrField(err))
			return
		case r := <-canceledCh:
			a.log.Warn("container listener channel canceled, reason: " + r)
			return
		}
	}
}
