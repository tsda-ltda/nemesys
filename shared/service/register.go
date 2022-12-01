package service

import (
	"context"
	"log"
	"time"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/uuid"
	"github.com/rabbitmq/amqp091-go"
)

func registerService(t Type) (n int, err error) {
	b, err := amqp.Encode(t)
	if err != nil {
		return n, err
	}

	amqpConn, err := amqp.Dial()
	if err != nil {
		return n, err
	}
	defer amqpConn.Close()

	p := models.NewAMQPPlumber()
	done := make(chan any)
	defer close(done)

	listenToRegisterReponse(amqpConn, p, done)

	ch, err := amqpConn.Channel()
	if err != nil {
		log.Fatalf("Fail to open socket channel, err: %s", err)
		return
	}

	err = ch.ExchangeDeclare(
		amqp.ExchangeServiceRegisterRequest, // name
		"fanout",                            // type
		true,                                // durable
		false,                               // autoDelete
		false,                               // internal
		false,                               // noWait
		nil,                                 // args
	)
	if err != nil {
		log.Fatalf("Fail to declare exchange, err: %s", err)
		return
	}

	for {
		log.Println("Sending service register request...")

		id, err := uuid.New()
		if err != nil {
			return n, err
		}
		err = ch.PublishWithContext(context.Background(),
			amqp.ExchangeServiceRegisterRequest, // exchange
			"",                                  // routing key
			false,                               // mandatory
			false,                               // immediate
			amqp091.Publishing{
				CorrelationId: id,
				Body:          b,
			},
		)
		if err != nil {
			return n, err
		}

		d, err := p.Listen(id, time.Second*5)
		if err != nil {
			log.Println("Service register request timeouted")
			continue
		}
		done <- nil

		err = amqp.Decode(d.Body, &n)
		if err != nil {
			return n, err
		}
		return n, nil
	}
}

func listenToRegisterReponse(conn *amqp091.Connection, p *models.AMQPPlumber, done <-chan any) {
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Fail to open socket channel, err: %s", err)
		return
	}

	err = ch.ExchangeDeclare(
		amqp.ExchangeServiceRegisterResponse, // name
		"fanout",                             // type
		true,                                 // durable
		false,                                // autoDelete
		false,                                // internal
		false,                                // noWait
		nil,                                  // args
	)
	if err != nil {
		log.Fatalf("Fail to declare exchange, err: %s", err)
		return
	}

	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // autoDelete
		true,  // exclusive
		false, // no wait
		nil,   // args
	)
	if err != nil {
		log.Fatalf("Fail to declare queue, err: %s", err)
		return
	}

	err = ch.QueueBind(
		q.Name,                               // name
		"",                                   // key
		amqp.ExchangeServiceRegisterResponse, // exchange
		false,                                // noWait
		nil,                                  // args
	)
	if err != nil {
		log.Fatalf("Fail to bind queue")
	}

	msgs, err := ch.Consume(
		q.Name, // name
		"",     // consumer
		true,   // autoAck
		false,  // exclusive
		false,  // noLocal
		false,  // noWait
		nil,    // args
	)
	if err != nil {
		log.Fatalf("Fail to consume channel, err: %s", err)
		return
	}

	closed, canceled := amqp.OnChannelCloseOrCancel(ch)
	go func() {

		for {
			select {
			case d := <-msgs:
				p.Send(d)
			case <-canceled:
				log.Fatalf("Unexpected channel cancelation")
				return
			case <-closed:
				log.Fatalf("Unexpected channel closed")
				return
			case <-done:
				return
			}
		}
	}()
}
