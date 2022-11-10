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

// basicMetricRequest is the basic information of a metric.
type basicMetricRequest struct {
	models.SNMPMetric
	// Type is the metric type.
	Type types.MetricType
}

// metricRequest is an extension of models.MetricRequest.
type metricRequest struct {
	basicMetricRequest
	// ContainerId is the metric's container identifier.
	ContainerId int32
	// Conn is the agent connection.
	Conn *Conn
}

// metricsRequest is an extension of models.MetricsRequest.
type metricsRequest struct {
	// ContainerId is the metric's container identifier.
	ContainerId int32
	// Conn is the agent connection.
	Conn *Conn
	// Metrics is the metrics.
	Metrics []basicMetricRequest
}

// metricDataPublisher receives a metric data request and fetch data,
// publishing it when available.
func (s *SNMPService) metricDataPublisher() {
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
		case r := <-s.metricDataReq:
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
					r.Info.Conn.Close()
					s.Log.Debug("fail to fetch data", logger.ErrField(err))
					p.Type = amqp.FromMessageType(amqp.Failed)
					return
				}

				// get first result
				pdu := pdus[0]

				// get metric type
				t := r.Info.Type
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
					s.Log.Debug("fail to parse pdu value to metric value", logger.ErrField(err))
					p.Type = amqp.FromMessageType(amqp.InvalidParse)
					return
				}

				// evaluate value
				v, err = s.evaluator.Evaluate(v, r.Info.Id, t)
				if err != nil {
					s.Log.Warn("fail to evaluate metric value", logger.ErrField(err))
					p.Type = amqp.FromMessageType(amqp.EvaluateFailed)
					return
				}

				res := models.MetricDataResponse{
					ContainerId: r.Info.ContainerId,
					MetricBasicDataReponse: models.MetricBasicDataReponse{
						Id:    req.Info.Id,
						Type:  t,
						Value: v,
					},
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

				s.Log.Debug("data published for metric: " + fmt.Sprint(req.Info.Id))
			}(ch, r)
		case <-s.stopDataPublisher:
			return
		}
	}
}

// metricsDataPublisher receives metrics data requests and fetch their data,
// publishing them when available.
func (s *SNMPService) metricsDataPublisher() {
	// open socket channel
	ch, err := s.amqp.Channel()
	if err != nil {
		s.Log.Panic("fail to open socket channel", logger.ErrField(err))
		return
	}
	defer ch.Close()

	// declare exchange
	err = ch.ExchangeDeclare(
		amqp.ExchangeMetricsData, // name
		"direct",                 // type
		true,                     // durable
		false,                    // auto-deleted
		false,                    // internal
		false,                    // no-wait
		nil,                      // arguments
	)
	if err != nil {
		s.Log.Panic("fail to declare exchange", logger.ErrField(err))
		return
	}

	for {
		select {
		case r := <-s.metricsDataReq:
			go func(ch *amqp091.Channel, req models.AMQPCorrelated[metricsRequest]) {
				ctx := context.Background()
				p := amqp091.Publishing{
					Headers:       amqp091.Table{"routing_key": req.RoutingKey},
					CorrelationId: req.CorrelationId,
				}

				// publish data
				defer func() {
					err = ch.PublishWithContext(ctx,
						amqp.ExchangeMetricsData, // exchange
						req.RoutingKey,           // routing key
						false,                    // mandatory
						false,                    // immediate
						p,
					)
					if err != nil {
						s.Log.Error("fail to publish data", logger.ErrField(err))
						return
					}
				}()

				// get OIDS
				oids := make([]string, len(req.Info.Metrics))
				for i, m := range req.Info.Metrics {
					oids[i] = m.OID
				}

				// fetch data
				pdus, err := s.Get(req.Info.Conn, oids)
				if err != nil {
					r.Info.Conn.Close()
					s.Log.Debug("fail to fetch data", logger.ErrField(err))
					p.Type = amqp.FromMessageType(amqp.Failed)
					return
				}

				// create response structure
				res := models.MetricsDataResponse{
					ContainerId: req.Info.ContainerId,
					Metrics:     make([]models.MetricBasicDataReponse, len(req.Info.Metrics)),
				}

				for i, pdu := range pdus {
					m := req.Info.Metrics[i]
					res.Metrics[i] = models.MetricBasicDataReponse{
						Id:     m.Id,
						Type:   m.Type,
						Value:  nil,
						Failed: false,
					}

					// parse SNMP response
					v, err := ParsePDU(pdu)
					if err != nil {
						s.Log.Debug("fail to parse PDU, name "+pdu.Name, logger.ErrField(err))
						res.Metrics[i].Failed = true
						continue
					}

					// parse value to metric type
					v, err = types.ParseValue(v, m.Type)
					if err != nil {
						s.Log.Warn("fail to parse SNMP value to metric value", logger.ErrField(err))
						res.Metrics[i].Failed = true
						continue
					}

					// evaluate value
					v, err = s.evaluator.Evaluate(v, m.Id, m.Type)
					if err != nil {
						s.Log.Warn("fail to evaluate value")
						res.Metrics[i].Failed = true
						continue
					}

					// set value
					res.Metrics[i].Value = v
				}

				// encode response
				bytes, err := amqp.Encode(res)
				if err != nil {
					s.Log.Error("fail to encode metrics data response", logger.ErrField(err))
					p.Type = amqp.FromMessageType(amqp.InternalError)
					return
				}

				p.Type = amqp.FromMessageType(amqp.OK)
				p.Body = bytes
				s.Log.Debug("metrics data published for container: " + fmt.Sprint(req.Info.ContainerId))
			}(ch, r)
		case <-s.stopDataPublisher:
			return
		}
	}
}
