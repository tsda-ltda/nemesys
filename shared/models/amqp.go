package models

import (
	"errors"
	"sync"
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
		channels: sync.Map{},
	}
}

type AMQPPlumber struct {
	channels sync.Map
}

// Send sends data to listener if exists.
func (p *AMQPPlumber) Send(delivery amqp091.Delivery) {
	ch, ok := p.channels.Load(delivery.CorrelationId)
	if !ok {
		return
	}
	ch.(chan amqp091.Delivery) <- delivery
}

// Listen creates and listen to a response.
func (p *AMQPPlumber) Listen(key string, timeout time.Duration) (amqp091.Delivery, error) {
	ch := make(chan amqp091.Delivery, 1)
	p.channels.Store(key, ch)
	defer p.channels.Delete(key)
	select {
	case res := <-ch:
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
