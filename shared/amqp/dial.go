package amqp

import (
	"fmt"

	"github.com/fernandotsda/nemesys/shared/env"
	amqp "github.com/rabbitmq/amqp091-go"
)

// Use current enviroment to connect to amqp server.
func Dial() (*amqp.Connection, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%s/", env.AMQPUsername, env.AMQPPassword, env.AMQPHost, env.AMQPPort)
	return amqp.Dial(url)
}
