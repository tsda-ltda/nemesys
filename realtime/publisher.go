package rts

import (
	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/rabbitmq/amqp091-go"
)

func (s *RTS) publishRTSMetricData(p amqp091.Publishing) {
	s.amqph.PublisherCh <- models.DetailedPublishing{
		Exchange:   amqp.ExchangeRTSMetricDataResponse,
		Publishing: p,
	}
	s.log.Debug("metric data published, id: <encoded>")
}
