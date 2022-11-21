package snmp

import (
	"strconv"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
)

// getMetricListener listen to metric data requests and send to
// the dataProducer.
func (s *SNMPService) getMetricListener() {
	msgs, err := s.amqph.Listen(amqp.QueueSNMPMetricDataRequest, amqp.ExchangeMetricDataRequest,
		models.ListenerOptions{Bind: models.QueueBindOptions{RoutingKey: "snmp"}},
	)
	if err != nil {
		s.log.Panic("fail to listen amqp messages", logger.ErrField(err))
		return
	}
	for d := range msgs {
		// decode message data
		var r models.MetricRequest
		err = amqp.Decode(d.Body, &r)
		if err != nil {
			s.log.Error("fail to unmarshal amqp message body", logger.ErrField(err))
			continue
		}
		s.log.Debug("get metric data request received, metric id: " + strconv.FormatInt(int64(r.MetricId), 10))

		// get gosnmp agent
		agent, err := s.getContainerAgent(r.ContainerId, r.ContainerType)
		if err != nil {
			s.log.Error("fail to get container config", logger.ErrField(err))
			continue
		}

		// get routing key on message header
		rk, err := amqp.GetRoutingKeyFromHeader(d.Headers)
		if err != nil {
			s.log.Error("fail to get routing key from header", logger.ErrField(err))
			continue
		}

		go s.getMetric(agent, r, d.CorrelationId, rk)
	}
}

// getMetricsListener listen to metrics data requests and send to
// the metricsDataProducer.
func (s *SNMPService) getMetricsListener() {
	msgs, err := s.amqph.Listen(amqp.QueueSNMPMetricsDataRequest, amqp.ExchangeMetricsDataRequest,
		models.ListenerOptions{Bind: models.QueueBindOptions{RoutingKey: "snmp"}},
	)
	if err != nil {
		s.log.Panic("fail to listen amqp messages", logger.ErrField(err))
		return
	}
	for d := range msgs {
		// decode message data
		var r models.MetricsRequest
		err = amqp.Decode(d.Body, &r)
		if err != nil {
			s.log.Error("fail to unmarshal amqp message body", logger.ErrField(err))
			continue
		}
		s.log.Debug("get metrics data request received, container id: " + strconv.FormatInt(int64(r.ContainerId), 10))

		// get gosnmp agent
		agent, err := s.getContainerAgent(r.ContainerId, r.ContainerType)
		if err != nil {
			s.log.Error("fail to get container config", logger.ErrField(err))
			continue
		}

		// get routing key on message header
		rk, err := amqp.GetRoutingKeyFromHeader(d.Headers)
		if err != nil {
			s.log.Error("fail to get routing key from header", logger.ErrField(err))
			continue
		}
		// fetch data asynchrounously
		go s.getMetrics(agent, r, d.CorrelationId, rk)
	}
}
