package amqph

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/logger"
)

func (a *Amqph) publisher() {
	// open socket channel
	ch, err := a.conn.Channel()
	if err != nil {
		a.log.Panic("Fail to open socket channel", logger.ErrField(err))
	}

	closed, canceled := amqp.OnChannelCloseOrCancel(ch)
	for {
		select {
		case r := <-a.PublisherCh:
			err = ch.PublishWithContext(context.Background(),
				r.Exchange,
				r.RoutingKey,
				r.Mandatory,
				r.Immediate,
				r.Publishing,
			)
			if err != nil {
				a.log.Error("Fail to publish message", logger.ErrField(err))
			}
		case err := <-closed:
			if err != nil {
				a.log.Panic("Publisher channel closed", logger.ErrField(err))
			}
			return
		case r := <-canceled:
			a.log.DPanic("Publisher channel canceled, reason: " + r)
			return
		}
	}

}
