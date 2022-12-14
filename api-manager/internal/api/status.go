package api

import (
	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/amqph"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/service"
)

func (api *API) servicesStatusListener() {
	var options amqph.ListenerOptions
	options.QueueDeclarationOptions.Exclusive = true
	options.QueueBindOptions.Exchange = amqp.ExchangeServicesStatus

	msgs, done := api.Amqph.Listen(options)
	for {
		select {
		case d := <-msgs:
			var status []service.ServiceStatus
			err := amqp.Decode(d.Body, &status)
			if err != nil {
				api.Log.Error("Fail to decode amqp message body", logger.ErrField(err))
				continue
			}
			api.servicesStatus = status
		case <-done:
			return
		case <-api.Done():
			return
		}
	}
}

func (api *API) GetServicesStatus() []service.ServiceStatus {
	return api.servicesStatus
}
