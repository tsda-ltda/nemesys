package amqph

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/rabbitmq/amqp091-go"
)

type Publish struct {
	Exchange   string
	RoutingKey string
	Mandatory  bool
	Immediate  bool
	Publishing amqp091.Publishing
}

func (a *Amqph) Publish(p Publish) {
	a.publisherCh <- p
}

func (a *Amqph) publisher() {
	ch, err := a.conn.Channel()
	if err != nil {
		a.log.Panic("Fail to open socket channel", logger.ErrField(err))
	}
	defer ch.Close()

	closed, canceled := amqp.OnChannelCloseOrCancel(ch)

	defer func() {
		close(closed)
		close(canceled)
	}()
	for {
		select {
		case p := <-a.publisherCh:
			err = ch.PublishWithContext(context.Background(),
				p.Exchange,
				p.RoutingKey,
				p.Mandatory,
				p.Immediate,
				p.Publishing,
			)
			if err != nil {
				a.log.Error("Fail to publish message", logger.ErrField(err))
			}
		case <-a.done():
			return
		case <-closed:
			return
		case <-canceled:
			return
		}
	}
}
