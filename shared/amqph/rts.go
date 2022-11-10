package amqph

import (
	"context"
	"errors"
	"time"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/uuid"
	"github.com/rabbitmq/amqp091-go"
)

var (
	ErrRequestTimeout = errors.New("request timeout")
)

// GetRTSData fetchs a metric data on real time service.
func (a *Amqph) GetRTSData(r models.MetricRequest) (d amqp091.Delivery, err error) {
	// encode request
	b, err := amqp.Encode(r)
	if err != nil {
		a.log.Error("fail to encode metric request", logger.ErrField(err))
		return d, err
	}

	// generate uuid
	uuid, err := uuid.New()
	if err != nil {
		a.log.Error("fail to create new uuid", logger.ErrField(err))
		return d, err
	}

	// send request
	a.rtsMetricDataRequestsCh <- amqp091.Publishing{
		Expiration:    "30000",
		Body:          b,
		CorrelationId: uuid,
	}

	// wait data
	d, err = a.plumber.Listen(uuid, time.Second*30)
	if err != nil {
		return d, ErrRequestTimeout
	}
	return d, nil
}

// ListenRTSMetricData listen to rts metric data.
func (a *Amqph) ListenRTSMetricData() {
	// open socket channel
	ch, err := a.conn.Channel()
	if err != nil {
		a.log.Panic("fail to open socket channel", logger.ErrField(err))
	}

	// declare data exchange
	err = ch.ExchangeDeclare(
		amqp.ExchangeRTSMetricData, // name
		"fanout",                   // type
		true,                       // durable
		false,                      // auto-deleted
		false,                      // internal
		false,                      // no-wait
		nil,                        // arguments
	)
	if err != nil {
		a.log.Panic("fail to declare exchange", logger.ErrField(err))
	}

	// declare queue
	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable'
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		a.log.Panic("fail to declare queue", logger.ErrField(err))
		return
	}

	// bind queue
	err = ch.QueueBind(
		q.Name,                     // queue name
		"",                         // routing key
		amqp.ExchangeRTSMetricData, // exchange
		false,                      // no-wait
		nil,                        // args
	)
	if err != nil {
		a.log.Panic("fail to bind queue", logger.ErrField(err))
		return
	}

	// consume messages
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		a.log.Panic("fail to consume messages", logger.ErrField(err))
	}

	// close and cancel channels
	closedCh := make(chan *amqp091.Error)
	canceledCh := make(chan string)
	ch.NotifyCancel(canceledCh)
	ch.NotifyClose(closedCh)

	for {
		select {
		case d := <-msgs:
			a.plumber.Send(d)
		case err := <-closedCh:
			a.log.Warn("RTS data channel closed", logger.ErrField(err))
			return
		case r := <-canceledCh:
			a.log.Warn("RTS data channel canceled, reason: " + r)
			return
		}
	}
}

// ListenRTSMetricDataRequests listen to rts metric data requets.
func (a *Amqph) ListenRTSMetricDataRequests() {
	// open socket channel
	ch, err := a.conn.Channel()
	if err != nil {
		a.log.Panic("fail to open socket channel", logger.ErrField(err))
	}

	// declare get data exchange
	err = ch.ExchangeDeclare(
		amqp.ExchangeRTSGetMetricData, // name
		"direct",                      // type
		true,                          // durable
		false,                         // auto-deleted
		false,                         // internal
		false,                         // no-wait
		nil,                           // arguments
	)
	if err != nil {
		a.log.Panic("fail to declare exchange", logger.ErrField(err))
	}

	// close and cancel channels
	closedCh := make(chan *amqp091.Error)
	canceledCh := make(chan string)
	ch.NotifyCancel(canceledCh)
	ch.NotifyClose(closedCh)

	for {
		select {
		case p := <-a.rtsMetricDataRequestsCh:
			err = ch.PublishWithContext(context.Background(),
				amqp.ExchangeRTSGetMetricData, // exchange
				"",                            // routing key
				false,                         // mandatory
				false,                         // immediate
				p,
			)
			if err != nil {
				a.log.Error("fail to publish rts metric data request", logger.ErrField(err))
			}
		case err := <-closedCh:
			a.log.Warn("RTS data channel closed", logger.ErrField(err))
			return
		case r := <-canceledCh:
			a.log.Warn("RTS data channel canceled, reason: " + r)
			return
		}
	}
}
