package logger

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/vmihailenco/msgpack/v5"
)

type AMQPLoggerWriter struct {
	ch *amqp.Channel
}

func (w *AMQPLoggerWriter) Write(p []byte) (n int, err error) {
	// marshal [][]byte
	b, err := msgpack.Marshal(p)
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
