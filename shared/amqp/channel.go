package amqp

import "github.com/rabbitmq/amqp091-go"

func OnChannelCloseOrCancel(ch *amqp091.Channel) (closed chan *amqp091.Error, canceled chan string) {
	closed = make(chan *amqp091.Error)
	canceled = make(chan string)
	ch.NotifyCancel(canceled)
	ch.NotifyClose(closed)
	return closed, canceled
}
