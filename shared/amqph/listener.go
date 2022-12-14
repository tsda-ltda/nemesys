package amqph

import (
	"time"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/rabbitmq/amqp091-go"
)

type ListenerOptions struct {
	QueueDeclarationOptions QueueDeclarationOptions
	QueueBindOptions        QueueBindOptions
	QueueConsumeOptions     QueueConsumeOptions
}

type QueueDeclarationOptions struct {
	Name              string
	Durable           bool
	DeletedWhenUnused bool
	Exclusive         bool
	NoWait            bool
	Arguments         amqp091.Table
}

type QueueBindOptions struct {
	RoutingKey string
	Exchange   string
	NoWait     bool
	Arguments  amqp091.Table
}

type QueueConsumeOptions struct {
	Consumer  string
	Exclusive bool
	NoLocal   bool
	NoWait    bool
	Arguments amqp091.Table
}

var listenerChannelReconnetionTimeout = time.Second * 10

func (a *Amqph) getDelivery(options ListenerOptions) (ch *amqp091.Channel, msgs <-chan amqp091.Delivery, err error) {
	ch, err = a.conn.Channel()
	if err != nil {
		return ch, msgs, err
	}

	q, err := ch.QueueDeclare(
		options.QueueDeclarationOptions.Name,
		options.QueueDeclarationOptions.Durable,
		options.QueueDeclarationOptions.DeletedWhenUnused,
		options.QueueDeclarationOptions.Exclusive,
		options.QueueDeclarationOptions.NoWait,
		options.QueueDeclarationOptions.Arguments,
	)
	if err != nil {
		return ch, msgs, err
	}

	err = ch.QueueBind(
		q.Name,
		options.QueueBindOptions.RoutingKey,
		options.QueueBindOptions.Exchange,
		options.QueueBindOptions.NoWait,
		options.QueueBindOptions.Arguments,
	)
	if err != nil {
		return ch, msgs, err
	}

	msgs, err = ch.Consume(
		q.Name,
		options.QueueConsumeOptions.Consumer,
		true, // auto-ack
		options.QueueConsumeOptions.Exclusive,
		options.QueueConsumeOptions.NoLocal,
		options.QueueConsumeOptions.NoWait,
		options.QueueConsumeOptions.Arguments,
	)
	return ch, msgs, err
}

func (a *Amqph) autoDelivery(options ListenerOptions, output chan amqp091.Delivery, done <-chan struct{}, noWait bool) {
	if !noWait {
		time.Sleep(listenerChannelReconnetionTimeout)
	}
	ch, msgs, err := a.getDelivery(options)
	if err != nil {
		a.log.Error("Fail to get amqp delivery", logger.ErrField(err))
		go a.autoDelivery(options, output, done, false)
		return
	}
	defer ch.Close()

	canceled, closed := amqp.OnChannelCloseOrCancel(ch)
	for {
		select {
		case d := <-msgs:
			output <- d
		case <-canceled:
			a.log.Error("Channel canceled, restarting auto delivery")
			go a.autoDelivery(options, output, done, false)
			return
		case <-closed:
			a.log.Error("Channel closed, restarting auto delivery")
			go a.autoDelivery(options, output, done, false)
			return
		case <-done:
			return
		}
	}
}

func (a *Amqph) Listen(options ListenerOptions) (msgs <-chan amqp091.Delivery, done <-chan struct{}) {
	output := make(chan amqp091.Delivery)
	go a.autoDelivery(options, output, a.done(), true)
	return output, a.done()
}
