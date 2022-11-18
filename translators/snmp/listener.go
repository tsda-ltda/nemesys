package snmp

import (
	"context"
	"strconv"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
)

func (s *SNMPService) containerListener() {
	for {
		select {
		case n := <-s.amqph.OnContainerUpdated():
			// close connection on container update
			c, ok := s.conns[n.Base.Id]
			if !ok {
				continue
			}
			c.Close()
		case id := <-s.amqph.OnContainerDeleted():
			// close connection on container delete
			c, ok := s.conns[id]
			if !ok {
				continue
			}
			c.Close()
		}
	}
}

func (s *SNMPService) metricListener() {
	for {
		select {
		case n := <-s.amqph.OnMetricUpdated():
			// close metric on update
			m, ok := s.metrics[n.Base.Id]
			if !ok {
				continue
			}
			m.Close()
		case pair := <-s.amqph.OnMetricDeleted():
			// stop metric pulling on metric delete
			m, ok := s.metrics[pair.Id]
			if !ok {
				continue
			}
			m.Close()
		}
	}
}

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
		ctx := context.Background()

		// decode message data
		var r models.MetricRequest
		err = amqp.Decode(d.Body, &r)
		if err != nil {
			s.log.Error("fail to unmarshal amqp message body", logger.ErrField(err))
			continue
		}
		s.log.Debug("get metric data request received, metric id: " + strconv.FormatInt(int64(r.MetricId), 10))

		// check if container connection exists
		c, ok := s.conns[r.ContainerId]
		if !ok {
			c, err = s.CreateContainerConnection(ctx, r.ContainerId, r.ContainerType)
			if err != nil {
				s.log.Error("fail to register agent", logger.ErrField(err))
				continue
			}
		} else {
			c.Reset()
		}

		// get metric oid
		metric, ok := s.metrics[r.MetricId]
		if !ok {
			metric, err = s.RegisterMetric(ctx, r, c.TTL)
			if err != nil {
				s.log.Error("fail to register metric", logger.ErrField(err))
				continue
			}
		} else {
			metric.Reset()
			metric.Type = r.MetricType
		}

		// get routing key on message header
		rk, err := amqp.GetRoutingKeyFromHeader(d.Headers)
		if err != nil {
			s.log.Error("fail to get routing key from header", logger.ErrField(err))
			continue
		}

		go s.GetMetric(c, r, metric.OID, d.CorrelationId, rk)
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
		ctx := context.Background()

		// decode message data
		var r models.MetricsRequest
		err = amqp.Decode(d.Body, &r)
		if err != nil {
			s.log.Error("fail to unmarshal amqp message body", logger.ErrField(err))
			continue
		}
		s.log.Debug("get metrics data request received, container id: " + strconv.FormatInt(int64(r.ContainerId), 10))

		// check if agent connection exists
		c, ok := s.conns[r.ContainerId]
		if !ok {
			c, err = s.CreateContainerConnection(ctx, r.ContainerId, r.ContainerType)
			if err != nil {
				s.log.Error("fail to register agent", logger.ErrField(err))
				continue
			}
		} else {
			c.Reset()
		}

		// get metrics full information
		metrics := make([]*Metric, 0, len(r.Metrics))
		metricsToRegister := []models.MetricBasicRequestInfo{}

		for _, m := range r.Metrics {
			// check if exists on local map
			metric, ok := s.metrics[m.Id]
			if ok {
				metric.Reset()
				metric.Type = m.Type
				metrics = append(metrics, metric)
				continue
			} else {
				metricsToRegister = append(metricsToRegister, m)
			}
		}

		if len(metricsToRegister) > 0 {
			// register metrics
			newMetrics, err := s.RegisterMetrics(ctx, metricsToRegister, r.ContainerType, c.TTL)
			if err != nil {
				s.log.Error("fail to register metrics", logger.ErrField(err))
				continue
			}
			metrics = append(metrics, newMetrics...)
		}

		// get routing key on message header
		rk, err := amqp.GetRoutingKeyFromHeader(d.Headers)
		if err != nil {
			s.log.Error("fail to get routing key from header", logger.ErrField(err))
			continue
		}

		// save oids
		oids := make([]string, len(metrics))
		for i, m := range metrics {
			oids[i] = m.OID
		}

		// fetch data asynchrounously
		go s.GetMetrics(c, r, oids, d.CorrelationId, rk)
	}
}
