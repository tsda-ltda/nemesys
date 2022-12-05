package manager

import (
	"time"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/service"
	"github.com/fernandotsda/nemesys/shared/uuid"
	"github.com/rabbitmq/amqp091-go"
)

func (s *ServiceManager) pingHandler() {
	emptyTime := time.Time{}
	ticker := time.NewTicker(s.pingInterval)
	for {
		select {
		case t := <-ticker.C:
			for i, serv := range s.services {
				go func(i int, serv service.ServiceStatus) {
					online := s.pingService(serv.Ident)

					if !serv.Online && online {
						s.log.Info("Service is online, ident: " + serv.Ident)
					}
					serv.Online = online
					serv.LastPing = t
					if online {
						serv.LostConnectionTime = emptyTime
					}
					if !online && serv.LostConnectionTime == emptyTime {
						serv.LostConnectionTime = t
						s.log.Warn("Service is offline, ident: " + serv.Ident)
					}
					s.services[i] = serv
				}(i, serv)
			}
			b, err := amqp.Encode(s.services)
			if err != nil {
				s.log.Error("Fail to encode amqp body", logger.ErrField(err))
				return
			}
			s.amqph.PublisherCh <- models.DetailedPublishing{
				Exchange: amqp.ExchangeServicesStatus,
				Publishing: amqp091.Publishing{
					Body: b,
				},
			}
		case <-s.Done():
			return
		}
	}
}

func (s *ServiceManager) pingService(serviceIdent string) (online bool) {
	pingId, err := uuid.New()
	if err != nil {
		s.log.Error("Fail to generate new uuid", logger.ErrField(err))
		return false
	}
	s.amqph.PublisherCh <- models.DetailedPublishing{
		Exchange:   amqp.ExchangeServicePing,
		RoutingKey: serviceIdent,
		Publishing: amqp091.Publishing{
			CorrelationId: pingId,
		},
	}
	_, err = s.pingPlumber.Listen(pingId, s.pingInterval)
	return err == nil
}

func (s *ServiceManager) pongHandler() {
	msgs, err := s.amqph.Listen("", amqp.ExchangeServicePong, models.ListenerOptions{
		Bind: models.QueueBindOptions{
			RoutingKey: "service-manager",
		},
	})
	if err != nil {
		s.log.Fatal("Fail to listen services pongs", logger.ErrField(err))
		return
	}
	for {
		select {
		case d := <-msgs:
			s.pingPlumber.Send(d)
		case <-s.Done():
			return
		}
	}
}
