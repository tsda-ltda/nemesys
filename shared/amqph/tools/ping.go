package tools

import (
	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/amqph"
	"github.com/rabbitmq/amqp091-go"
)

func ServicePing(a *amqph.Amqph, serviceIdent string) {
	var options amqph.ListenerOptions
	options.QueueDeclarationOptions.Exclusive = true
	options.QueueBindOptions.RoutingKey = serviceIdent
	options.QueueBindOptions.Exchange = amqp.ExchangeServicePing

	msgs, done := a.Listen(options)
	for {
		select {
		case d := <-msgs:
			a.Publish(amqph.Publish{
				Exchange:   amqp.ExchangeServicePong,
				RoutingKey: "service-manager",
				Publishing: amqp091.Publishing{
					CorrelationId: d.CorrelationId,
				},
			})
		case <-done:
			return
		}
	}
}
