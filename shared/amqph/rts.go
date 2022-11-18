package amqph

import (
	"time"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/uuid"
	"github.com/rabbitmq/amqp091-go"
)

var rtsMetricDataListenerInitialized = false

// GetRTSData fetchs a metric data on real time service.
func (a *Amqph) GetRTSData(r models.MetricRequest) (d amqp091.Delivery, err error) {
	a.listenRTSMetricData()

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
	a.PublisherCh <- models.DetailedPublishing{
		Exchange: amqp.ExchangeRTSMetricDataRequest,
		Publishing: amqp091.Publishing{
			Expiration:    "30000",
			Body:          b,
			CorrelationId: uuid,
		},
	}

	// wait data
	d, err = a.plumber.Listen(uuid, time.Second*30)
	if err != nil {
		return d, ErrRequestTimeout
	}
	return d, nil
}

// ListenRTSMetricData listen to rts metric data.
func (a *Amqph) listenRTSMetricData() {
	if rtsMetricDataListenerInitialized {
		return
	}

	go func() {
		msgs, err := a.Listen("", amqp.ExchangeRTSMetricDataResponse)
		if err != nil {
			a.log.Panic("fail to listen amqp messages", logger.ErrField(err))
			return
		}
		for d := range msgs {
			a.plumber.Send(d)
		}
	}()
	rtsMetricDataListenerInitialized = true
}
