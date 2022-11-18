package amqph

import (
	"errors"

	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/rabbitmq/amqp091-go"
)

func (a *Amqph) Listen(queueName string, exchange string, options ...models.ListenerOptions) (msgs <-chan amqp091.Delivery, err error) {
	var option *models.ListenerOptions
	if len(options) == 0 {
		option = &models.ListenerOptions{}
	} else if len(options) == 1 {
		option = &options[0]
	} else {
		return msgs, errors.New("options may not more than 1 element")
	}

	// open socket channel
	ch, err := a.conn.Channel()
	if err != nil {
		return msgs, err
	}

	// declare queue
	q, err := ch.QueueDeclare(
		queueName,           // name
		option.Durable,      // durable'
		option.AutoDelete,   // delete when unused
		!option.NoExclusive, // exclusive
		option.NoWait,       // no-wait
		option.Args,         // arguments
	)
	if err != nil {
		return msgs, err
	}

	// bind queue
	err = ch.QueueBind(
		q.Name,                 // queue name
		option.Bind.RoutingKey, // routing key
		exchange,               // exchange
		option.Bind.NoWait,     // no-wait
		option.Bind.Args,       // args
	)
	if err != nil {
		return msgs, err
	}

	// consume messages
	msgs, err = ch.Consume(
		queueName,                 // queue
		option.Consume.Consumer,   // consumer
		!option.Consume.ManualAck, // auto-ack
		option.Consume.Exclusive,  // exclusive
		option.Consume.NoLocal,    // no-local
		option.Consume.NoWait,     // no-wait
		option.Consume.Args,       // args
	)
	if err != nil {
		return msgs, err
	}
	return msgs, nil
}
