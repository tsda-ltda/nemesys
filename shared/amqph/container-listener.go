package amqph

import (
	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/rabbitmq/amqp091-go"
)

// ContainerListener receives notifications and calls a callback.
func (a *Amqph) ContainerListener() {
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
		qCreate.Name,                        // queue name
		"",                                  // routing key
		amqp.ExchangeNotifyCreatedContainer, // exchange
		false,
		nil,
	)
	if err != nil {
		a.log.Panic("fail to bind queue")
	}

	err = ch.QueueBind(
		qUpdate.Name,                        // queue name
		"",                                  // routing key
		amqp.ExchangeNotifyUpdatedContainer, // exchange
		false,
		nil,
	)
	if err != nil {
		a.log.Panic("fail to bind queue")
	}

	err = ch.QueueBind(
		qDelete.Name,                        // queue name
		"",                                  // routing key
		amqp.ExchangeNotifyDeletedContainer, // exchange
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
			var n ContainerNotification
			err = amqp.Decode(d.Body, &d)
			if err != nil {
				a.log.Error("fail to decode container notification message body", logger.ErrField(err))
				continue
			}
			a.Notifications.Publish(ContainerCreated, n)
		case d := <-msgsUpdate:
			var n ContainerNotification
			err = amqp.Decode(d.Body, &n)
			if err != nil {
				a.log.Error("fail to decode container notification message body", logger.ErrField(err))
				continue
			}
			a.Notifications.Publish(ContainerUpdated, n)
		case d := <-msgsDelete:
			var id int32
			err = amqp.Decode(d.Body, &id)
			if err != nil {
				a.log.Error("fail to decode container notification message body", logger.ErrField(err))
				continue
			}
			a.Notifications.Publish(ContainerDeleted, id)
		case err := <-closedCh:
			a.log.Warn("container listener channel closed", logger.ErrField(err))
			return
		case r := <-canceledCh:
			a.log.Warn("container listener channel canceled, reason: " + r)
			return
		}
	}
}
