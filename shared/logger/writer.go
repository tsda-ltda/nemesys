package logger

import (
	"context"

	_amqp "github.com/fernandotsda/nemesys/shared/amqp"
	amqp "github.com/rabbitmq/amqp091-go"
)

type AMQPLoggerWriter struct {
	ch *amqp.Channel
}

func (w *AMQPLoggerWriter) Write(p []byte) (n int, err error) {
	// econde bytes
	b, err := _amqp.Encode(p)
	if err != nil {
		return 0, err
	}

	err = w.ch.PublishWithContext(context.Background(),
		"logs", // exchange
		"",     // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        b,
		},
	)

	return 0, err
}
