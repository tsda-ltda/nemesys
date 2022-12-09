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
		a.log.Panic("Fail to open socket channel")
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
			a.log.Panic("Fail declare exchange ("+name+").", logger.ErrField(err))
		}
	}

	declare(amqp.ExchangeContainerCreated, "fanout", true, false, false, false, nil)
	declare(amqp.ExchangeContainerUpdated, "fanout", true, false, false, false, nil)
	declare(amqp.ExchangeContainerDeleted, "fanout", true, false, false, false, nil)
	declare(amqp.ExchangeMetricCreated, "fanout", true, false, false, false, nil)
	declare(amqp.ExchangeMetricUpdated, "fanout", true, false, false, false, nil)
	declare(amqp.ExchangeMetricDeleted, "fanout", true, false, false, false, nil)
	declare(amqp.ExchangeDataPolicyDeleted, "fanout", true, false, false, false, nil)
	declare(amqp.ExchangeServiceLogs, "fanout", true, false, false, false, nil)
	declare(amqp.ExchangeServicesStatus, "fanout", true, false, false, false, nil)
	declare(amqp.ExchangeServiceRegisterReq, "fanout", true, false, false, false, nil)
	declare(amqp.ExchangeServiceRegisterRes, "fanout", true, false, false, false, nil)
	declare(amqp.ExchangeServiceUnregister, "fanout", true, false, false, false, nil)
	declare(amqp.ExchangeCheckMetricsAlarm, "fanout", true, false, false, false, nil)
	declare(amqp.ExchangeCheckMetricAlarm, "fanout", true, false, false, false, nil)
	declare(amqp.ExchangeMetricsAlarmed, "fanout", true, false, false, false, nil)
	declare(amqp.ExchangeMetricAlarmed, "fanout", true, false, false, false, nil)

	declare(amqp.ExchangeServicePing, "direct", true, false, false, false, nil)
	declare(amqp.ExchangeServicePong, "direct", true, false, false, false, nil)
	declare(amqp.ExchangeMetricDataReq, "direct", true, false, false, false, nil)
	declare(amqp.ExchangeMetricDataRes, "direct", true, false, false, false, nil)
	declare(amqp.ExchangeMetricsDataReq, "direct", true, false, false, false, nil)
	declare(amqp.ExchangeMetricsDataRes, "direct", true, false, false, false, nil)
}
