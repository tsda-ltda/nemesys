package rts

import (
	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/rabbitmq/amqp091-go"
)

func (s *RTS) publishRTSMetricData(p amqp091.Publishing, rk string) {
	s.amqph.PublisherCh <- models.DetailedPublishing{
		Exchange:   amqp.ExchangeMetricDataRes,
		RoutingKey: rk,
		Publishing: p,
	}
	s.log.Debug("Metric data published")
}
