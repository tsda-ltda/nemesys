package snmp

import (
	"context"
	"fmt"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/rabbitmq/amqp091-go"
	"github.com/vmihailenco/msgpack/v5"
)

type DataReq struct {
	Conn     *Conn
	OIDS     []string
	Metadata msgpack.RawMessage
}

// dataProducer receives data request through SNMPService.dataReq and fetch data asynchronously,
// publishing it in the amqp server.
func (s *SNMPService) dataProducer() {
	// open socket channel
	ch, err := s.amqp.Channel()
	if err != nil {
		s.Log.Panic("fail to open socket channel", logger.ErrField(err))
		return
	}
	defer ch.Close()

	// declare exchange
	err = ch.ExchangeDeclare(
		amqp.ExchangeSNMPData, // name
		"fanout",              // type
		true,                  // durable
		false,                 // auto-deleted
		false,                 // internal
		false,                 // no-wait
		nil,                   // arguments
	)
	if err != nil {
		s.Log.Panic("fail to declare exchange", logger.ErrField(err))
		return
	}

	for {
		select {
		case req := <-s.dataReq:
			go func(ch *amqp091.Channel, req DataReq) {
				ctx := context.Background()

				// fetch data
				r, err := s.Get(req.Conn, req.OIDS)
				if err != nil {
					s.Log.Warn("fail to get data", logger.ErrField(err))
					return
				}

				// encode data
				data, err := msgpack.Marshal(r)
				if err != nil {
					s.Log.Error("fail to marshal data response", logger.ErrField(err))
					return
				}

				// encode message
				bytes, err := msgpack.Marshal(models.AMQPMessage{
					Type:     amqp.MTUntyped,
					Metadata: req.Metadata,
					Data:     data,
				})
				if err != nil {
					s.Log.Error("fail to marshal amqp message", logger.ErrField(err))
					return
				}

				err = ch.PublishWithContext(ctx,
					amqp.ExchangeSNMPData, // exchange
					"",                    // routing key
					false,                 // mandatory
					false,                 // immediate
					amqp091.Publishing{
						ContentType: "application/msgpack",
						Body:        bytes,
					})
				if err != nil {
					s.Log.Error("fail to publish data", logger.ErrField(err))
					return
				}
				s.Log.Debug(fmt.Sprintf("data published for %s:%d", req.Conn.Agent.Target, req.Conn.Agent.Port))
			}(ch, req)
		case <-s.Done:
			return
		}
	}
}
