package amqph

import (
	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/rabbitmq/amqp091-go"
)

func (a *Amqph) declareExchages() {
	// open socket channel
	ch, err := a.conn.Channel()
	if err != nil {
		a.log.Panic("fail to open socket channel")
	}
	defer ch.Close()

	// helper function
	declare := func(name string, kind string, durable bool, autoDelete bool, internal bool, noWait bool, args amqp091.Table) {
		err = ch.ExchangeDeclare(
			name,
			kind,
			durable,
			autoDelete,
			internal,
			noWait,
			args,
		)
		if err != nil {
			a.log.Panic("fail declare exchange ("+name+").", logger.ErrField(err))
		}
	}

	declare(amqp.ExchangeDataPolicyDeleted, "fanout", true, false, false, false, nil)

	declare(amqp.ExchangeContainerCreated, "fanout", true, false, false, false, nil)
	declare(amqp.ExchangeContainerUpdated, "fanout", true, false, false, false, nil)
	declare(amqp.ExchangeContainerDeleted, "fanout", true, false, false, false, nil)

	declare(amqp.ExchangeMetricCreated, "fanout", true, false, false, false, nil)
	declare(amqp.ExchangeMetricUpdated, "fanout", true, false, false, false, nil)
	declare(amqp.ExchangeMetricDeleted, "fanout", true, false, false, false, nil)

	declare(amqp.ExchangeMetricDataRequest, "direct", true, false, false, false, nil)
	declare(amqp.ExchangeMetricDataResponse, "direct", true, false, false, false, nil)
	declare(amqp.ExchangeMetricsDataRequest, "direct", true, false, false, false, nil)
	declare(amqp.ExchangeMetricsDataResponse, "direct", true, false, false, false, nil)

	declare(amqp.ExchangeRTSMetricDataRequest, "direct", true, false, false, false, nil)
	declare(amqp.ExchangeRTSMetricDataResponse, "direct", true, false, false, false, nil)
}
