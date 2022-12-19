package snmp

import (
	"strconv"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/rabbitmq/amqp091-go"
)

func (s *SNMP) getMetricDataHandler(d amqp091.Delivery) {
	var r models.MetricRequest
	err := amqp.Decode(d.Body, &r)
	if err != nil {
		s.log.Error("Fail to unmarshal amqp message body", logger.ErrField(err))
		return
	}
	s.log.Debug("Get metric data request received, metric id: " + strconv.FormatInt(int64(r.MetricId), 10))

	agent, err := s.getContainerAgent(r.ContainerId, r.ContainerType)
	if err != nil {
		s.log.Error("Fail to get container config", logger.ErrField(err))
		return
	}

	rk, err := amqp.GetRoutingKeyFromHeader(d.Headers)
	if err != nil {
		s.log.Error("Fail to get routing key from header", logger.ErrField(err))
		return
	}

	s.fetchMetricData(agent, r, d.CorrelationId, rk)
}

func (s *SNMP) getMetricsDataHandler(d amqp091.Delivery) {
	var r models.MetricsRequest
	err := amqp.Decode(d.Body, &r)
	if err != nil {
		s.log.Error("Fail to unmarshal amqp message body", logger.ErrField(err))
		return
	}
	s.log.Debug("Get metrics data request received, container id: " + strconv.FormatInt(int64(r.ContainerId), 10))

	agent, err := s.getContainerAgent(r.ContainerId, r.ContainerType)
	if err != nil {
		s.log.Error("Fail to get container config", logger.ErrField(err))
		return
	}

	rk, err := amqp.GetRoutingKeyFromHeader(d.Headers)
	if err != nil {
		s.log.Error("Fail to get routing key from header", logger.ErrField(err))
		return
	}
	s.fetchMetricsData(agent, r, d.CorrelationId, rk)
}
