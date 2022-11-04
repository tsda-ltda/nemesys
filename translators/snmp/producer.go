package snmp

import (
	"context"
	"fmt"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/types"
	"github.com/rabbitmq/amqp091-go"
)

// metricRequest is an extension of models.MetricRequest.
type metricRequest struct {
	models.MetricRequest
	Conn *Conn
	OID  string
}

// dataPublisher receives data request fetch data asynchronously,
// publishing it in the amqp server.
func (s *SNMPService) dataPublisher() {
	// open socket channel
	ch, err := s.amqp.Channel()
	if err != nil {
		s.Log.Panic("fail to open socket channel", logger.ErrField(err))
		return
	}
	defer ch.Close()

	// declare exchange
	err = ch.ExchangeDeclare(
		amqp.ExchangeMetricData, // name
		"direct",                // type
		true,                    // durable
		false,                   // auto-deleted
		false,                   // internal
		false,                   // no-wait
		nil,                     // arguments
	)
	if err != nil {
		s.Log.Panic("fail to declare exchange", logger.ErrField(err))
		return
	}

	for {
		select {
		case r := <-s.singleDataReq:
			go func(ch *amqp091.Channel, req models.AMQPCorrelated[metricRequest]) {
				ctx := context.Background()
				p := amqp091.Publishing{
					Headers:       amqp091.Table{"routing_key": req.RoutingKey},
					CorrelationId: req.CorrelationId,
				}

				// publish data
				defer func() {
					err = ch.PublishWithContext(ctx,
						amqp.ExchangeMetricData, // exchange
						req.RoutingKey,          // routing key
						false,                   // mandatory
						false,                   // immediate
						p,
					)
					if err != nil {
						s.Log.Error("fail to publish data", logger.ErrField(err))
						return
					}
				}()

				// fetch data
				pdus, err := s.Get(req.Info.Conn, []string{req.Info.OID})
				if err != nil {
					s.Log.Debug("fail to fetch data", logger.ErrField(err))
					p.Type = amqp.FromMessageType(amqp.Failed)
					return
				}

				// get first result
				pdu := pdus[0]

				// get metric type
				t := r.Info.MetricType
				if t == types.MTUnknown {
					t = types.ParseAsn1BER(byte(pdu.Type))
				}

				// parse raw value
				v, err := ParsePDU(pdu)
				if err != nil {
					s.Log.Debug("fail to parse pdu value", logger.ErrField(err))
					p.Type = amqp.FromMessageType(amqp.Failed)
					return
				}

				// parse value to metric type
				v, err = types.ParseValue(v, t)
				if err != nil {
					s.Log.Error("fail to parse pdu value to metric value", logger.ErrField(err))
					p.Type = amqp.FromMessageType(amqp.InternalError)
					return
				}

				// evaluate value
				v, err = s.evaluator.Evaluate(v, r.Info.MetricId, t)
				if err != nil {
					s.Log.Warn("fail to evaluate metric value", logger.ErrField(err))
					p.Type = amqp.FromMessageType(amqp.Failed)
					return
				}

				res := models.MetricDataResponse{
					ContainerId: r.Info.ContainerId,
					MetricId:    r.Info.MetricId,
					Value:       v,
					MetricType:  t,
				}

				// encode data
				bytes, err := amqp.Encode(res)
				if err != nil {
					s.Log.Error("fail to marshal data response", logger.ErrField(err))
					p.Type = amqp.FromMessageType(amqp.InternalError)
					return
				}

				// set data
				p.Type = amqp.FromMessageType(amqp.OK)
				p.Body = bytes

				s.Log.Debug("data published for metric: " + fmt.Sprint(req.Info.MetricId))
			}(ch, r)
		case <-s.stopDataPublisher:
			return
		}
	}
}
