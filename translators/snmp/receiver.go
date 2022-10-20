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

// getReceiver is responsable for receiver get data messages and redirect to
// the dataProducer.
func (s *SNMPService) getReceiver() {
	// create socket channel
	ch, err := s.amqp.Channel()
	if err != nil {
		s.Log.Panic("fail to open amqp socket channel", logger.ErrField(err))
		return
	}
	defer ch.Close()

	// declare exchange
	err = ch.ExchangeDeclare(
		amqp.ExchangeSNMPGet, // name
		"fanout",             // type
		true,                 // durable
		false,                // auto-deleted
		false,                // internal
		false,                // no-wait
		nil,                  // arguments
	)
	if err != nil {
		s.Log.Panic("fail to declare exchange", logger.ErrField(err))
		return
	}

	// declare get queue
	qGet, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		s.Log.Panic("fail to declare queue", logger.ErrField(err))
		return
	}

	// declare registration required queue
	qRegisterRequired, err := ch.QueueDeclare(
		amqp.QueueSNMPConnRegistRequired, // name
		true,                             // durable
		false,                            // delete when unused
		false,                            // exclusive
		false,                            // no-wait
		nil,                              // arguments
	)
	if err != nil {
		s.Log.Panic("fail to declare queue", logger.ErrField(err))
		return
	}

	// bind get queue
	err = ch.QueueBind(
		qGet.Name,            // queue name
		"",                   // routing key
		amqp.ExchangeSNMPGet, // exchange
		false,                // no-wait
		nil,                  // arguments
	)
	if err != nil {
		s.Log.Panic("fail to bind queue", logger.ErrField(err))
		return
	}

	// consume get queue
	msgs, err := ch.Consume(
		qGet.Name, // queue
		"",        // consumer
		true,      // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		s.Log.Panic("fail to consume queue", logger.ErrField(err))
		return
	}

	for {
		select {
		case d := <-msgs:
			ctx := context.Background()

			// decode amqp message
			var m models.AMQPMessage
			err = msgpack.Unmarshal(d.Body, &m)
			if err != nil {
				s.Log.Error("fail to unmarhsal amqp message body", logger.ErrField(err))
				continue
			}

			// decode message data
			var info models.SNMPGetMetrics
			err = msgpack.Unmarshal(m.Data, &info)
			if err != nil {
				s.Log.Error("fail to unmarshal amqp message data", logger.ErrField(err))
				continue
			}

			// check if agent connection exists
			c := s.GetConn(info.Target, info.Port)
			if c == nil {
				// publish to a work queue
				ch.PublishWithContext(ctx,
					"",                     // exchange
					qRegisterRequired.Name, // routing key
					false,                  // mandatory
					false,                  // immediate
					amqp091.Publishing{
						DeliveryMode: amqp091.Transient,
						ContentType:  "application/msgpack",
						Body:         d.Body, // just retransmit incomming data
					})
				s.Log.Debug(fmt.Sprintf("no connection registered for data request, addr: %s:%d", info.Target, info.Port))
				continue
			}

			// resert ttl ticker
			c.Reset()

			// add data request
			s.dataReq <- DataReq{
				Conn:     c,
				OIDS:     info.OIDS,
				Metadata: m.Metadata,
			}
			s.Log.Debug(fmt.Sprintf("new data request for %s:%d added", info.Target, info.Port))
		case <-s.Done:
			return
		}
	}
}

// registerConnReceiver is responsable for register snmp agent's connections.
// If connection already exists will overwrite it.
func (s *SNMPService) registerConnReceiver() {
	// create socket channel
	ch, err := s.amqp.Channel()
	if err != nil {
		s.Log.Panic("fail to open amqp socket channel", logger.ErrField(err))
		return
	}
	defer ch.Close()

	// declare exchange
	err = ch.ExchangeDeclare(
		amqp.ExchangeSNMPConnRegister, // name
		"fanout",                      // type
		true,                          // durable
		false,                         // auto-deleted
		false,                         // internal
		false,                         // no-wait
		nil,                           // arguments
	)
	if err != nil {
		s.Log.Panic("fail to declare exchange", logger.ErrField(err))
		return
	}

	// declare queue
	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		s.Log.Panic("fail to declare queue", logger.ErrField(err))
		return
	}

	// bind queue
	err = ch.QueueBind(
		q.Name,                        // queue name
		"",                            // routing key
		amqp.ExchangeSNMPConnRegister, // exchange
		false,                         // no-wait
		nil,                           // arguments
	)
	if err != nil {
		s.Log.Panic("fail to bind queue", logger.ErrField(err))
		return
	}

	// consume get queue
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
		s.Log.Panic("fail to consume queue", logger.ErrField(err))
		return
	}

	for {
		select {
		case d := <-msgs:
			// decode amqp message
			var m models.AMQPMessage
			err = msgpack.Unmarshal(d.Body, &m)
			if err != nil {
				s.Log.Error("fail to unmarhsal amqp message body", logger.ErrField(err))
				continue
			}

			// decode message data
			var conf models.SNMPAgentConfig
			err = msgpack.Unmarshal(m.Data, &conf)
			if err != nil {
				s.Log.Error("fail to unmarshal amqp message data", logger.ErrField(err))
				continue
			}
			s.Log.Debug(fmt.Sprintf("registering agent connection for %s:%d", conf.Target, conf.Port))

			// check if agent connection exists
			c := s.GetConn(conf.Target, conf.Port)
			if c != nil {
				c.Close()
			}

			// register connection
			err = s.RegisterConn(conf)
			if err != nil {
				s.Log.Warn("fail to register agent connection")
				return
			}

			continue
		case <-s.Done:
			return
		}
	}
}
