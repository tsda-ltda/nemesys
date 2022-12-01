package manager

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/service"
	jsoniter "github.com/json-iterator/go"
	"github.com/rabbitmq/amqp091-go"
)

func (s *ServiceManager) logListener() {
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	msgs, err := s.amqph.Listen("", amqp.ExchangeServiceLogs)
	if err != nil {
		s.log.Fatal("Fail to listen service unregister requests", logger.ErrField(err))
		return
	}
	for {
		select {
		case d := <-msgs:
			var log map[string]any
			err := json.Unmarshal(d.Body, &log)
			if err != nil {
				continue
			}
			s.influxClient.WriteLog(context.Background(), log)
		case <-s.Done():
			return
		}
	}
}

func (s *ServiceManager) registryListener() {
	register, err := s.amqph.Listen("", amqp.ExchangeServiceRegisterRequest)
	if err != nil {
		s.log.Fatal("Fail to listen service register requests", logger.ErrField(err))
		return
	}
	unregister, err := s.amqph.Listen("", amqp.ExchangeServiceUnregister)
	if err != nil {
		s.log.Fatal("Fail to listen service unregister requests", logger.ErrField(err))
		return
	}
	for {
		select {
		case d := <-register:
			var t service.Type
			err := amqp.Decode(d.Body, &t)
			if err != nil {
				s.log.Error("Fail to decode amqp body", logger.ErrField(err))
				continue
			}

			service := s.newService(t)
			b, err := amqp.Encode(service.Number)
			if err != nil {
				s.log.Error("Fail to encode body", logger.ErrField(err))
				continue
			}
			s.amqph.PublisherCh <- models.DetailedPublishing{
				Exchange: amqp.ExchangeServiceRegisterResponse,
				Publishing: amqp091.Publishing{
					CorrelationId: d.CorrelationId,
					Body:          b,
				},
			}
			s.log.Info("New service registrated, service ident: " + service.Ident)
		case d := <-unregister:
			var ident string
			err := amqp.Decode(d.Body, &ident)
			if err != nil {
				s.log.Error("Fail to decode amqp body", logger.ErrField(err))
				continue
			}
			founded := false
			newServices := make([]service.ServiceStatus, len(s.services)-1)
			for _, serv := range s.services {
				if serv.Ident == ident {
					founded = true
					switch serv.Type {
					case service.Alarm:
						alarmN.Release(serv.Number)
					case service.RTS:
						rtsN.Release(serv.Number)
					case service.APIManager:
						apiManagerN.Release(serv.Number)
					case service.DHS:
						dhsN.Release(serv.Number)
					case service.SNMP:
						snmpN.Release(serv.Number)
					case service.WS:
						wsN.Release(serv.Number)
					}
					continue
				}
				newServices = append(newServices, serv)
			}
			if !founded {
				s.log.Error("Received a unregister service request but service was not registered, service ident: " + ident)
				continue
			}
			s.services = newServices
		case <-s.Done():
			return
		}
	}
}
