package rts

import (
	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/amqph"
	"github.com/fernandotsda/nemesys/shared/models"
)

func (s *RTS) onDataPolicyDeleted(id int16) {
	for _, cp := range s.pulling {
		cp.Close()
	}
}

func (s *RTS) onContainerUpdated(base models.BaseContainer, protocol any) {
	c, ok := s.pulling[base.Id]
	if !ok {
		return
	}
	c.Close()
}

func (s *RTS) onContainerDeleted(id int32) {
	c, ok := s.pulling[id]
	if !ok {
		return
	}
	c.Close()
}

func (s *RTS) onMetricUpdated(base models.BaseMetric, protocol any) {
	c, ok := s.pulling[base.ContainerId]
	if !ok {
		return
	}
	m, ok := c.Metrics[base.Id]
	if !ok {
		return
	}
	m.Stop()
}

func (s *RTS) onMetricDeleted(containerId int32, id int64) {
	cp, ok := s.pulling[containerId]
	if !ok {
		return
	}
	m, ok := cp.Metrics[id]
	if !ok {
		return
	}
	m.Stop()
}

func (s *RTS) metricDataRequestListener() {
	var options amqph.ListenerOptions
	options.QueueDeclarationOptions.Name = amqp.QueueRTSMetricDataReq
	options.QueueBindOptions.Exchange = amqp.ExchangeMetricDataReq
	options.QueueBindOptions.RoutingKey = "rts"

	msgs, done := s.amqph.Listen(options)
	for {
		select {
		case d := <-msgs:
			go s.metricDataRequestHandler(d)
		case <-done:
			return
		}
	}
}

// metricDataListener listen to metric data response, using a unique routing key,
// to resolve rts data requests.
func (s *RTS) metricDataListener() {
	var options amqph.ListenerOptions
	options.QueueDeclarationOptions.Exclusive = true
	options.QueueBindOptions.Exchange = amqp.ExchangeMetricDataRes
	options.QueueBindOptions.RoutingKey = s.GetServiceIdent()

	msgs, done := s.amqph.Listen(options)
	for {
		select {
		case d := <-msgs:
			go s.metricDataHandler(d)
		case <-done:
			return
		}
	}
}

// globalMetricDataListener listen to metric data response, using rts routing key,
// to resolve incoming data that are not rts requests responses.
func (s *RTS) globalMetricDataListener() {
	var options amqph.ListenerOptions
	options.QueueDeclarationOptions.Name = amqp.QueueRTSMetricData
	options.QueueBindOptions.Exchange = amqp.ExchangeMetricDataRes
	options.QueueBindOptions.RoutingKey = "rts"

	msgs, done := s.amqph.Listen(options)
	for {
		select {
		case d := <-msgs:
			go s.metricDataHandler(d)
		case <-done:
			return
		}
	}
}

// metricsDataListener listen to metrics data response.
func (s *RTS) metricsDataListener() {
	var options amqph.ListenerOptions
	options.QueueDeclarationOptions.Exclusive = true
	options.QueueBindOptions.Exchange = amqp.ExchangeMetricsDataRes
	options.QueueBindOptions.RoutingKey = s.GetServiceIdent()

	msgs, done := s.amqph.Listen(options)
	for {
		select {
		case d := <-msgs:
			go s.metricsDataHandler(d)
		case <-done:
			return
		}
	}
}
