package snmp

import (
	"math"
	"strconv"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/amqph"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/types"
	"github.com/gosnmp/gosnmp"
	"github.com/rabbitmq/amqp091-go"
)

func (s *SNMP) fetchMetricData(agent models.SNMPv2cAgent, request models.MetricRequest, correlationId string, routingKey string) {
	p := amqp091.Publishing{
		Headers:       amqp.RouteHeader(routingKey),
		CorrelationId: correlationId,
	}

	metricsRes, failed, err := s.getSNMPMetricsData(agent, models.MetricsRequest{
		ContainerId:   request.ContainerId,
		ContainerType: request.ContainerType,
		Metrics: []models.MetricBasicRequestInfo{{
			Id:           request.MetricId,
			Type:         request.MetricType,
			DataPolicyId: request.DataPolicyId,
		}},
	})
	if err != nil {
		if failed {
			p.Type = amqp.FromMessageType(amqp.Failed)
		} else {
			p.Type = amqp.FromMessageType(amqp.InternalError)
		}
	}

	if len(metricsRes.Metrics) < 1 {
		p.Type = amqp.GetMessage(amqp.InternalError)
		s.log.Error("Expected at least 1 metric data response got: " + strconv.Itoa(len(metricsRes.Metrics)))
	} else {
		response := models.MetricDataResponse{
			ContainerId:            metricsRes.ContainerId,
			MetricBasicDataReponse: metricsRes.Metrics[0],
		}
		if !metricsRes.Metrics[1].Failed {
			p.Type = amqp.FromMessageType(amqp.OK)
			b, err := amqp.Encode(response)
			if err != nil {
				p.Type = amqp.FromMessageType(amqp.InternalError)
				s.log.Error("Fail to encode amqp body", logger.ErrField(err))
			}
			p.Body = b
		}
	}

	s.amqph.Publish(amqph.Publish{
		Exchange:   amqp.ExchangeMetricDataRes,
		RoutingKey: routingKey,
		Publishing: p,
	})
	s.log.Debug("Metric data published, metric id: " + strconv.FormatInt(request.MetricId, 10))

	if types.IsNonFlex(request.ContainerType) && p.Type == amqp.FromMessageType(amqp.OK) {
		s.amqph.Publish(amqph.Publish{
			Exchange:   amqp.ExchangeCheckMetricAlarm,
			Publishing: p,
		})
		s.log.Debug("Metric data sent to alarm validation")
	}
}

func (s *SNMP) fetchMetricsData(agent models.SNMPv2cAgent, request models.MetricsRequest, correlationId string, routingKey string) {
	p := amqp091.Publishing{
		Headers:       amqp.RouteHeader(routingKey),
		CorrelationId: correlationId,
	}

	metricsRes, failed, err := s.getSNMPMetricsData(agent, request)
	if err != nil {
		if failed {
			p.Type = amqp.FromMessageType(amqp.Failed)
		} else {
			p.Type = amqp.FromMessageType(amqp.InternalError)
		}
	}

	b, err := amqp.Encode(metricsRes)
	if err != nil {
		p.Type = amqp.FromMessageType(amqp.InternalError)
		s.log.Error("Fail to encode amqp body", logger.ErrField(err))
	}
	p.Body = b
	p.Type = amqp.FromMessageType(amqp.OK)

	s.amqph.Publish(amqph.Publish{
		Exchange:   amqp.ExchangeMetricsDataRes,
		RoutingKey: routingKey,
		Publishing: p,
	})
	s.log.Debug("Metrics data published, container id: " + strconv.FormatInt(int64(request.ContainerId), 10))

	if types.IsNonFlex(request.ContainerType) && p.Type == amqp.FromMessageType(amqp.OK) {
		s.amqph.Publish(amqph.Publish{
			Exchange:   amqp.ExchangeCheckMetricsAlarm,
			Publishing: p,
		})

		s.log.Debug("Metrics data sent to alarm validation")
	}
}

func (s *SNMP) getSNMPMetricsData(agent models.SNMPv2cAgent, request models.MetricsRequest) (response models.MetricsDataResponse, fetchFailed bool, err error) {
	gosnmp := &gosnmp.GoSNMP{
		Target:    agent.Target,
		Port:      agent.Port,
		Transport: agent.Transport,
		Community: agent.Community,
		Version:   agent.Version,
		Timeout:   agent.Timeout,
		Retries:   agent.Retries,
		MaxOids:   agent.MaxOids,
	}

	err = gosnmp.Connect()
	if err != nil {
		s.log.Error("Fail to connect agent", logger.ErrField(err))
		return response, false, err
	}
	defer gosnmp.Conn.Close()

	metrics, err := s.getSNMPMetrics(models.MetricsRequest{
		ContainerId:   request.ContainerId,
		ContainerType: request.ContainerType,
		Metrics:       request.Metrics,
	})
	if err != nil {
		s.log.Error("Fail to get snmp metrics", logger.ErrField(err))
		return response, false, err
	}

	oids := make([]string, 0, len(metrics)*3)
	for _, m := range metrics {
		oids = append(oids, m.OID)
	}

	// flex legacy alarm and category oids
	var extraOidsIndexs []int
	if request.ContainerType == types.CTFlexLegacy {
		extraOidsIndexs = make([]int, 0, len(metrics)*2)
		for i, oid := range oids {
			alarm, err := getFlexLegacyAlarmOID(oid)
			if err != nil {
				continue
			}
			trap, err := getFlexLegacyCategoryOID(oid)
			if err != nil {
				continue
			}
			extraOidsIndexs = append(extraOidsIndexs, i)
			oids = append(oids, alarm, trap)
		}
	}

	pdus, err := s.get(gosnmp, oids)
	if err != nil {
		s.log.Debug("Fail to fetch data", logger.ErrField(err))
		return response, true, err
	}

	res := models.MetricsDataResponse{
		ContainerId: request.ContainerId,
		Metrics:     make([]models.MetricBasicDataReponse, len(oids)),
	}

	for i := 0; i < len(oids)-len(extraOidsIndexs)*2; i++ {
		pdu := pdus[i]
		r := request.Metrics[i]
		res.Metrics[i] = models.MetricBasicDataReponse{
			Id:           r.Id,
			Type:         r.Type,
			Value:        nil,
			DataPolicyId: r.DataPolicyId,
			Failed:       false,
		}

		v, err := parsePDU(pdu)
		if err != nil {
			s.log.Debug("Fail to parse PDU, name "+pdu.Name, logger.ErrField(err))
			res.Metrics[i].Failed = true
			continue
		}

		v, err = types.ParseValue(v, r.Type)
		if err != nil {
			s.log.Debug("Fail to parse SNMP value to metric value", logger.ErrField(err))
			res.Metrics[i].Failed = true
			continue
		}

		v, err = s.evaluator.Evaluate(v, r.Id, r.Type)
		if err != nil {
			s.log.Debug("Fail to evaluate value")
			res.Metrics[i].Failed = true
			continue
		}

		res.Metrics[i].Value = v
	}

	if len(extraOidsIndexs) > 0 {
		var j int
		alarms := make([]flexLegacyAlarm, len(extraOidsIndexs))

		for i := len(oids) - (len(extraOidsIndexs) * 2); i < len(oids); i += 2 {
			v, err := parsePDU(pdus[i])
			if err != nil {
				s.log.Debug("Fail to parse pdu", logger.ErrField(err))
				continue
			}

			v, err = types.ParseValue(v, types.MTBool)
			if err != nil {
				s.log.Error("Fail to parse alarm state to boolean", logger.ErrField(err))
				continue
			}

			alarmed, ok := v.(bool)
			if !ok {
				s.log.Error("Fail to make boolean type asseption on alarm state")
				continue
			}

			if !alarmed {
				continue
			}

			v, err = parsePDU(pdus[i+1])
			if err != nil {
				s.log.Debug("Fail to parse pdu", logger.ErrField(err))
				continue
			}

			v, err = types.ParseValue(v, types.MTInt)
			if err != nil {
				s.log.Error("Fail to parse trap category to integer", logger.ErrField(err))
				continue
			}

			trapCategoryId, ok := v.(int64)
			if !ok {
				s.log.Error("Fail to make int type asseption on trap category id")
				continue
			}

			res := res.Metrics[extraOidsIndexs[j]]
			if res.Failed {
				continue
			}

			alarms[j] = flexLegacyAlarm{
				MetricId: res.Id,
				Value:    res.Value,
				TrapId:   int16(trapCategoryId),
			}
		}

		if len(alarms) > 0 {
			go s.notifyAlarms(request.ContainerId, alarms)
		}
	}

	return res, false, nil
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

	res = make([]gosnmp.SnmpPDU, 0, len(oids))

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
