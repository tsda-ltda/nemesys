package snmp

import (
	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/amqph"
)

func (s *SNMP) getMetricDataListener() {
	var options amqph.ListenerOptions
	options.QueueDeclarationOptions.Name = amqp.QueueSNMPMetricDataReq
	options.QueueBindOptions.Exchange = amqp.ExchangeMetricDataReq
	options.QueueBindOptions.RoutingKey = "snmp"

	msgs, done := s.amqph.Listen(options)
	for {
		select {
		case d := <-msgs:
			go s.getMetricDataHandler(d)
		case <-done:
			return
		}
	}
}

func (s *SNMP) getMetricsDataListener() {
	var options amqph.ListenerOptions
	options.QueueDeclarationOptions.Name = amqp.QueueSNMPMetricsDataReq
	options.QueueBindOptions.Exchange = amqp.ExchangeMetricsDataReq
	options.QueueBindOptions.RoutingKey = "snmp"

	msgs, done := s.amqph.Listen(options)
	for {
		select {
		case d := <-msgs:
			go s.getMetricsDataHandler(d)
		case <-done:
			return
		}
	}
}
