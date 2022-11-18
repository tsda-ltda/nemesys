package amqph

import (
	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/rabbitmq/amqp091-go"
)

// RequestMetricData sends a metric data request to translators. The data can be received on the OnMetricData handler.
func (a *Amqph) RequestMetricData(req models.MetricRequest, serviceName string) error {
	b, err := amqp.Encode(req)
	if err != nil {
		return err
	}
	a.PublisherCh <- models.DetailedPublishing{
		Exchange:   amqp.ExchangeMetricDataRequest,
		RoutingKey: amqp.GetDataRoutingKey(req.ContainerType),
		Publishing: amqp091.Publishing{
			Headers:    amqp.RouteHeader(serviceName),
			Expiration: amqp.DefaultExp,
			Body:       b,
		},
	}
	return nil
}
