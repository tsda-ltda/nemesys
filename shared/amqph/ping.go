package amqph

import (
	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/rabbitmq/amqp091-go"
)

func (a *Amqph) pingHandler() {
	msgs, err := a.Listen("", amqp.ExchangeServicePing, models.ListenerOptions{
		Bind: models.QueueBindOptions{
			RoutingKey: a.serviceIdent,
		},
	})
	if err != nil {
		a.log.Fatal("Fail to listen to ping messages", logger.ErrField(err))
		return
	}

	for {
		select {
		case d := <-msgs:
			a.PublisherCh <- models.DetailedPublishing{
				Exchange: amqp.ExchangeServicePong,
				Publishing: amqp091.Publishing{
					CorrelationId: d.CorrelationId,
				},
			}
		case <-a.conn.NotifyClose(make(chan *amqp091.Error)):
			return
		}
	}
}
