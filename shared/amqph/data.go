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
	routingKey, err := amqp.GetDataRoutingKey(req.ContainerType)
	if err != nil {
		return nil
	}
	a.PublisherCh <- models.DetailedPublishing{
		Exchange:   amqp.ExchangeMetricDataRequest,
		RoutingKey: routingKey,
		Publishing: amqp091.Publishing{
			Headers:    amqp.RouteHeader(serviceName),
			Expiration: amqp.DefaultExp,
			Body:       b,
		},
	}
	return nil
}
