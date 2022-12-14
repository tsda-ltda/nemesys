package snmp

import (
	"strconv"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/amqph"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
)

func (s *SNMP) getMetricListener() {
	var options amqph.ListenerOptions
	options.QueueDeclarationOptions.Name = amqp.QueueSNMPMetricDataReq
	options.QueueBindOptions.Exchange = amqp.ExchangeMetricDataReq
	options.QueueBindOptions.RoutingKey = "snmp"

	msgs, done := s.amqph.Listen(options)
	for {
		select {
		case d := <-msgs:
			var r models.MetricRequest
			err := amqp.Decode(d.Body, &r)
			if err != nil {
				s.log.Error("Fail to unmarshal amqp message body", logger.ErrField(err))
				continue
			}
			s.log.Debug("Get metric data request received, metric id: " + strconv.FormatInt(int64(r.MetricId), 10))

			agent, err := s.getContainerAgent(r.ContainerId, r.ContainerType)
			if err != nil {
				s.log.Error("Fail to get container config", logger.ErrField(err))
				continue
			}

			rk, err := amqp.GetRoutingKeyFromHeader(d.Headers)
			if err != nil {
				s.log.Error("Fail to get routing key from header", logger.ErrField(err))
				continue
			}

			go s.fetchMetricData(agent, r, d.CorrelationId, rk)
		case <-done:
			return
		}
	}
}

func (s *SNMP) getMetricsListener() {
	var options amqph.ListenerOptions
	options.QueueDeclarationOptions.Name = amqp.QueueSNMPMetricsDataReq
	options.QueueBindOptions.Exchange = amqp.ExchangeMetricsDataReq
	options.QueueBindOptions.RoutingKey = "snmp"

	msgs, done := s.amqph.Listen(options)
	for {
		select {
		case d := <-msgs:
			var r models.MetricsRequest
			err := amqp.Decode(d.Body, &r)
			if err != nil {
				s.log.Error("Fail to unmarshal amqp message body", logger.ErrField(err))
				continue
			}
			s.log.Debug("Get metrics data request received, container id: " + strconv.FormatInt(int64(r.ContainerId), 10))

			agent, err := s.getContainerAgent(r.ContainerId, r.ContainerType)
			if err != nil {
				s.log.Error("Fail to get container config", logger.ErrField(err))
				continue
			}

			rk, err := amqp.GetRoutingKeyFromHeader(d.Headers)
			if err != nil {
				s.log.Error("Fail to get routing key from header", logger.ErrField(err))
				continue
			}
			go s.fetchMetricsData(agent, r, d.CorrelationId, rk)
		case <-done:
			return
		}
	}
}
