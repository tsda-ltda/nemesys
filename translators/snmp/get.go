package snmp

import (
	"fmt"
	"math"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/types"
	"github.com/gosnmp/gosnmp"
	"github.com/rabbitmq/amqp091-go"
)

func (s *SNMPService) GetMetric(conn *ContainerConn, request models.MetricRequest, oid string, correlationId string, routingKey string) {
	p := amqp091.Publishing{
		Headers:       amqp.RouteHeader(routingKey),
		CorrelationId: correlationId,
	}

	// publish data
	defer func() {
		s.amqph.PublisherCh <- models.DetailedPublishing{
			Exchange:   amqp.ExchangeMetricDataResponse,
			RoutingKey: routingKey,
			Publishing: p,
		}
	}()

	// fetch data
	pdus, err := s.Get(conn, []string{oid})
	if err != nil {
		conn.Close()
		s.log.Debug("fail to fetch data", logger.ErrField(err))
		p.Type = amqp.FromMessageType(amqp.Failed)
		return
	}

	// get first result
	pdu := pdus[0]

	// get metric type
	t := request.MetricType
	if t == types.MTUnknown {
		t = types.ParseAsn1BER(byte(pdu.Type))
	}

	// parse raw value
	v, err := ParsePDU(pdu)
	if err != nil {
		s.log.Debug("fail to parse pdu value", logger.ErrField(err))
		p.Type = amqp.FromMessageType(amqp.Failed)
		return
	}

	// parse value to metric type
	v, err = types.ParseValue(v, t)
	if err != nil {
		s.log.Debug("fail to parse pdu value to metric value", logger.ErrField(err))
		p.Type = amqp.FromMessageType(amqp.InvalidParse)
		return
	}

	// evaluate value
	v, err = s.evaluator.Evaluate(v, request.MetricId, t)
	if err != nil {
		s.log.Warn("fail to evaluate metric value", logger.ErrField(err))
		p.Type = amqp.FromMessageType(amqp.EvaluateFailed)
		return
	}

	res := models.MetricDataResponse{
		ContainerId: request.ContainerId,
		MetricBasicDataReponse: models.MetricBasicDataReponse{
			Id:    request.MetricId,
			Type:  t,
			Value: v,
		},
	}

	// encode data
	bytes, err := amqp.Encode(res)
	if err != nil {
		s.log.Error("fail to marshal data response", logger.ErrField(err))
		p.Type = amqp.FromMessageType(amqp.InternalError)
		return
	}

	// set data
	p.Type = amqp.FromMessageType(amqp.OK)
	p.Body = bytes

	s.log.Debug("data published for metric: " + fmt.Sprint(request.MetricId))
}

func (s *SNMPService) GetMetrics(conn *ContainerConn, request models.MetricsRequest, oids []string, correlationId string, routingKey string) {
	p := amqp091.Publishing{
		Headers:       amqp.RouteHeader(routingKey),
		CorrelationId: correlationId,
	}

	// publish data
	defer func() {
		s.amqph.PublisherCh <- models.DetailedPublishing{
			Exchange:   amqp.ExchangeMetricsDataResponse,
			RoutingKey: routingKey,
			Publishing: p,
		}
	}()

	// fetch data
	pdus, err := s.Get(conn, oids)
	if err != nil {
		conn.Close()
		s.log.Debug("fail to fetch data", logger.ErrField(err))
		p.Type = amqp.FromMessageType(amqp.Failed)
		return
	}

	// create response structure
	res := models.MetricsDataResponse{
		ContainerId: request.ContainerId,
		Metrics:     make([]models.MetricBasicDataReponse, len(oids)),
	}

	for i, pdu := range pdus {
		for _, r := range request.Metrics {
			m := s.metrics[r.Id]
			if m.OID != pdu.Name {
				continue
			}

			res.Metrics[i] = models.MetricBasicDataReponse{
				Id:     m.Id,
				Type:   m.Type,
				Value:  nil,
				Failed: false,
			}

			// parse SNMP response
			v, err := ParsePDU(pdu)
			if err != nil {
				s.log.Debug("fail to parse PDU, name "+pdu.Name, logger.ErrField(err))
				res.Metrics[i].Failed = true
				continue
			}

			// parse value to metric type
			v, err = types.ParseValue(v, m.Type)
			if err != nil {
				s.log.Warn("fail to parse SNMP value to metric value", logger.ErrField(err))
				res.Metrics[i].Failed = true
				continue
			}

			// evaluate value
			v, err = s.evaluator.Evaluate(v, m.Id, m.Type)
			if err != nil {
				s.log.Warn("fail to evaluate value")
				res.Metrics[i].Failed = true
				continue
			}

			// set value
			res.Metrics[i].Value = v
			break
		}
	}

	// encode response
	bytes, err := amqp.Encode(res)
	if err != nil {
		s.log.Error("fail to encode metrics data response", logger.ErrField(err))
		p.Type = amqp.FromMessageType(amqp.InternalError)
		return
	}

	p.Type = amqp.FromMessageType(amqp.OK)
	p.Body = bytes
	s.log.Debug("metrics data published for container: " + fmt.Sprint(request.ContainerId))
}

// Get fetch the OIDs's values. Returns an error only if an error is returned fo the SNMP Get request.
func (s *SNMPService) Get(c *ContainerConn, oids []string) (res []gosnmp.SnmpPDU, err error) {
	// get agent
	a := c.Agent

	// oids buffer
	var oidsBuff []string
	if len(oids) >= a.MaxOids {
		oidsBuff = make([]string, a.MaxOids)
	} else {
		oidsBuff = make([]string, len(oids))
	}

	res = []gosnmp.SnmpPDU{}

	var i int
	for k := 0; k < int(math.Ceil(float64(len(oids))/float64(a.MaxOids))); k++ {
		// recalculate buffer
		r := len(oids) - k*a.MaxOids
		if r <= cap(oidsBuff) {
			oidsBuff = make([]string, r)
		}

		// get oids
		for j := 0; j < len(oidsBuff); j++ {
			oidsBuff[j] = oids[i]
			i++
		}

		// make request
		_res, _err := a.Get(oidsBuff)
		err = _err
		if err != nil {
			return res, err
		}

		// save response
		res = append(res, _res.Variables...)
	}
	return res, err
}
