package amqph

import (
	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/rabbitmq/amqp091-go"
)

// NotifyDataPolicyDeleted notifies that a data policy was deleted.
func (a *Amqph) NotifyDataPolicyDeleted(id int16) error {
	b, err := amqp.Encode(id)
	if err != nil {
		return err
	}

	a.PublisherCh <- models.DetailedPublishing{
		Exchange: amqp.ExchangeDataPolicyDeleted,
		Publishing: amqp091.Publishing{
			Expiration: amqp.DefaultExp,
			Body:       b,
		},
	}
	return nil
}

func (a *Amqph) OnDataPolicyDeleted(queue ...string) <-chan int16 {
	var q string
	if len(queue) > 0 {
		q = queue[0]
	}
	delivery := make(chan int16)
	go func() {
		msgs, err := a.Listen(q, amqp.ExchangeDataPolicyDeleted)
		if err != nil {
			a.log.Panic("Fail to listen amqp messages", logger.ErrField(err))
			return
		}
		for d := range msgs {
			var id int16
			err = amqp.Decode(d.Body, &id)
			if err != nil {
				a.log.Error("Fail to decode delivery body", logger.ErrField(err))
				continue
			}
			delivery <- id
		}
	}()
	return delivery
}
