package snmp

import (
	"context"
	"fmt"
	"math"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/types"
	"github.com/gosnmp/gosnmp"
	"github.com/rabbitmq/amqp091-go"
)

func (s *SNMP) getMetric(agent models.SNMPAgent, request models.MetricRequest, correlationId string, routingKey string) {
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

	gosnmp := &gosnmp.GoSNMP{
		Context:            context.Background(),
		Target:             agent.Target,
		Port:               agent.Port,
		Transport:          agent.Transport,
		Community:          agent.Community,
		Version:            agent.Version,
		Timeout:            agent.Timeout,
		Retries:            agent.Retries,
		MaxOids:            agent.MaxOids,
		MsgFlags:           agent.MsgFlags,
		SecurityModel:      agent.SecurityModel,
		SecurityParameters: agent.SecurityParameters,
		ContextEngineID:    agent.ContextEngineID,
		ContextName:        agent.ContextName,
	}

	// open connection
	err := gosnmp.Connect()
	if err != nil {
		s.log.Debug("fail to connect agent", logger.ErrField(err))
		return
	}
	defer gosnmp.Conn.Close()

	// get metric
	metrics, err := s.getSNMPMetrics(models.MetricsRequest{
		ContainerId:   request.ContainerId,
		ContainerType: request.ContainerType,
		Metrics: []models.MetricBasicRequestInfo{{
			Id:           request.MetricId,
			Type:         request.MetricType,
			DataPolicyId: request.DataPolicyId,
		}},
	})
	if err != nil {
		s.log.Debug("fail to get snmp metrics", logger.ErrField(err))
		return
	}
	oid := metrics[0].OID

	// fetch data
	pdus, err := s.get(gosnmp, []string{oid})
	if err != nil {
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
			Id:           request.MetricId,
			Type:         t,
			Value:        v,
			DataPolicyId: request.DataPolicyId,
			Failed:       false,
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

func (s *SNMP) getMetrics(agent models.SNMPAgent, request models.MetricsRequest, correlationId string, routingKey string) {
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

	gosnmp := &gosnmp.GoSNMP{
		Context:            context.Background(),
		Target:             agent.Target,
		Port:               agent.Port,
		Transport:          agent.Transport,
		Community:          agent.Community,
		Version:            agent.Version,
		Timeout:            agent.Timeout,
		Retries:            agent.Retries,
		MaxOids:            agent.MaxOids,
		MsgFlags:           agent.MsgFlags,
		SecurityModel:      agent.SecurityModel,
		SecurityParameters: agent.SecurityParameters,
		ContextEngineID:    agent.ContextEngineID,
		ContextName:        agent.ContextName,
	}

	// connect
	err := gosnmp.Connect()
	if err != nil {
		s.log.Error("fail to connect agent", logger.ErrField(err))
		return
	}
	defer gosnmp.Conn.Close()

	// get metric
	metrics, err := s.getSNMPMetrics(models.MetricsRequest{
		ContainerId:   request.ContainerId,
		ContainerType: request.ContainerType,
		Metrics:       request.Metrics,
	})
	if err != nil {
		s.log.Error("fail to get snmp metrics", logger.ErrField(err))
	}

	// get oids
	oids := make([]string, len(metrics))
	for i, m := range metrics {
		oids[i] = m.OID
	}

	// fetch data
	pdus, err := s.get(gosnmp, oids)
	if err != nil {
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
			for _, _m := range metrics {
				if _m.Id == r.Id && pdu.Name == _m.OID {
					res.Metrics[i] = models.MetricBasicDataReponse{
						Id:           r.Id,
						Type:         r.Type,
						Value:        nil,
						DataPolicyId: r.DataPolicyId,
						Failed:       false,
					}

					// parse SNMP response
					v, err := ParsePDU(pdu)
					if err != nil {
						s.log.Debug("fail to parse PDU, name "+pdu.Name, logger.ErrField(err))
						res.Metrics[i].Failed = true
						continue
					}

					// parse value to metric type
					v, err = types.ParseValue(v, r.Type)
					if err != nil {
						s.log.Warn("fail to parse SNMP value to metric value", logger.ErrField(err))
						res.Metrics[i].Failed = true
						continue
					}

					// evaluate value
					v, err = s.evaluator.Evaluate(v, r.Id, r.Type)
					if err != nil {
						s.log.Warn("fail to evaluate value")
						res.Metrics[i].Failed = true
						continue
					}

					// set value
					res.Metrics[i].Value = v
					break
				}
				break
			}
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
func (s *SNMP) get(agent *gosnmp.GoSNMP, oids []string) (res []gosnmp.SnmpPDU, err error) {
	// oids buffer
	var oidsBuff []string
	if len(oids) >= agent.MaxOids {
		oidsBuff = make([]string, agent.MaxOids)
	} else {
		oidsBuff = make([]string, len(oids))
	}

	res = []gosnmp.SnmpPDU{}

	var i int
	for k := 0; k < int(math.Ceil(float64(len(oids))/float64(agent.MaxOids))); k++ {
		// recalculate buffer
		r := len(oids) - k*agent.MaxOids
		if r <= cap(oidsBuff) {
			oidsBuff = make([]string, r)
		}

		// get oids
		for j := 0; j < len(oidsBuff); j++ {
			oidsBuff[j] = oids[i]
			i++
		}

		// make request
		_res, _err := agent.Get(oidsBuff)
		err = _err
		if err != nil {
			return res, err
		}

		// save response
		res = append(res, _res.Variables...)
	}
	return res, err
}
