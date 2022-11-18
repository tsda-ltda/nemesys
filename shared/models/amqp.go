package models

import (
	"errors"
	"time"

	amqp091 "github.com/rabbitmq/amqp091-go"
)

type AMQPCorrelated[T any] struct {
	CorrelationId string
	RoutingKey    string
	Info          T
}

func NewAMQPPlumber() *AMQPPlumber {
	return &AMQPPlumber{
		channels: map[string]chan amqp091.Delivery{},
	}
}

type AMQPPlumber struct {
	channels map[string]chan amqp091.Delivery
}

// Send sends data to listener if exists.
func (p *AMQPPlumber) Send(delivery amqp091.Delivery) {
	ch, ok := p.channels[delivery.CorrelationId]
	if !ok {
		return
	}
	ch <- delivery
}

// Listen creates and listen to a response.
func (p *AMQPPlumber) Listen(key string, timeout time.Duration) (amqp091.Delivery, error) {
	p.channels[key] = make(chan amqp091.Delivery)
	defer close(p.channels[key])
	defer delete(p.channels, key)
	select {
	case res := <-p.channels[key]:
		return res, nil
	case <-time.After(timeout):
		return amqp091.Delivery{}, errors.New("response timeout")
	}
}

type DetailedPublishing struct {
	Exchange   string
	RoutingKey string
	Mandatory  bool
	Immediate  bool
	Publishing amqp091.Publishing
}

type QueueConsumeOptions struct {
	Consumer  string
	ManualAck bool
	Exclusive bool
	NoLocal   bool
	NoWait    bool
	Args      amqp091.Table
}

type QueueBindOptions struct {
	RoutingKey string
	NoWait     bool
	Args       amqp091.Table
}

type ListenerOptions struct {
	Durable     bool
	AutoDelete  bool
	NoExclusive bool
	NoWait      bool
	Args        amqp091.Table
	Consume     QueueConsumeOptions
	Bind        QueueBindOptions
}
