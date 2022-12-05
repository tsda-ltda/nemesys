package snmp

import (
	"strconv"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
)

func (s *SNMP) getMetricListener() {
	msgs, err := s.amqph.Listen(amqp.QueueSNMPMetricDataReq, amqp.ExchangeMetricDataReq,
		models.ListenerOptions{Bind: models.QueueBindOptions{RoutingKey: "snmp"}},
	)
	if err != nil {
		s.log.Panic("Fail to listen amqp messages", logger.ErrField(err))
		return
	}
	for {
		select {
		case d := <-msgs:
			var r models.MetricRequest
			err = amqp.Decode(d.Body, &r)
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

			go s.getSNMPv2cMetric(agent, r, d.CorrelationId, rk)
		case <-s.Done():
			return
		}
	}
}

func (s *SNMP) getMetricsListener() {
	msgs, err := s.amqph.Listen(amqp.QueueSNMPMetricsDataReq, amqp.ExchangeMetricsDataReq,
		models.ListenerOptions{Bind: models.QueueBindOptions{RoutingKey: "snmp"}},
	)
	if err != nil {
		s.log.Panic("Fail to listen amqp messages", logger.ErrField(err))
		return
	}
	for {
		select {
		case d := <-msgs:
			var r models.MetricsRequest
			err = amqp.Decode(d.Body, &r)
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
			go s.getMetrics(agent, r, d.CorrelationId, rk)
		case <-s.Done():
			return
		}
	}
}
