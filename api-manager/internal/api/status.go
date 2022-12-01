package api

import (
	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/service"
)

func (api *API) servicesStatusListener() {
	msgs, err := api.Amqph.Listen("", amqp.ExchangeServicesStatus)
	if err != nil {
		api.Log.Fatal("Fail to listen to services status", logger.ErrField(err))
		return
	}
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
		case <-api.Done():
			return
		}
	}
}

func (api *API) GetServicesStatus() []service.ServiceStatus {
	return api.servicesStatus
}
