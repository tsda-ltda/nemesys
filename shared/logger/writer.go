package logger

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/rabbitmq/amqp091-go"
)

type AMQPLoggerWriter struct {
	ch *amqp091.Channel
}

func (w *AMQPLoggerWriter) Write(p []byte) (n int, err error) {
	err = w.ch.PublishWithContext(context.Background(),
		amqp.ExchangeServiceLogs, // exchange
		"",                       // routing key
		false,                    // mandatory
		false,                    // immediate
		amqp091.Publishing{
			Body: p,
		},
	)

	return 0, err
}
