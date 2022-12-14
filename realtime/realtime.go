package rts

import (
	stdlog "log"
	"strconv"
	"sync"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/amqph"
	t "github.com/fernandotsda/nemesys/shared/amqph/tools"
	"github.com/fernandotsda/nemesys/shared/cache"
	"github.com/fernandotsda/nemesys/shared/env"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/pg"
	"github.com/fernandotsda/nemesys/shared/service"
	"github.com/rabbitmq/amqp091-go"
)

// Real Time Service
type RTS struct {
	service.Tools
	// cache is the cache handler
	cache *cache.Cache
	// pg is the postgresql handler.
	pg *pg.PG
	// amqp is the amqp connection.
	amqp *amqp091.Connection
	// amqph is the amqp handler for common taks.
	amqph *amqph.Amqph
	// log is the logger handler.
	log *logger.Logger
	// plumber is the custom amqp message plumber
	// for deal with request responses cases.
	plumber *models.AMQPPlumber
	// muStartPulling is the mutex to add metrics on pulling.
	muStartPulling sync.Mutex
	// pulling is the map of containers pulling.
	pulling map[int32]*ContainerPulling
	// pendingMetricDataRequest is a map of pending data requests.
	pendingMetricDataRequest map[string]models.RTSMetricConfig
}

func New(serviceNumber int) service.Service {
	tools := service.NewTools(service.RTS, serviceNumber)
	amqpConn, err := amqp.Dial()
	if err != nil {
		stdlog.Panicf("Fail to dial with amqp server, err: %s", err)
		return nil
	}

	log, err := logger.New(
		amqpConn,
		logger.Config{
			Service:        tools.ServiceIdent,
			ConsoleLevel:   logger.ParseLevelEnv(env.LogConsoleLevelRTS),
			BroadcastLevel: logger.ParseLevelEnv(env.LogBroadcastLevelRTS),
		},
	)
	if err != nil {
		stdlog.Panicf("Fail to create logger, err: %s", err)
	}
	log.Info("Connected to amqp server")

	publishers, err := strconv.Atoi(env.RTSAMQPPublishers)
	if err != nil {
		log.Fatal("Fail to parse env.RTSAMQPPublishers", logger.ErrField(err))
		return nil
	}

	amqph := amqph.New(amqph.Config{
		Log:        log,
		Conn:       amqpConn,
		Publishers: publishers,
	})
	go t.ServicePing(amqph, tools.ServiceIdent)

	return &RTS{
		Tools:                    tools,
		log:                      log,
		pg:                       pg.New(),
		amqp:                     amqpConn,
		amqph:                    amqph,
		cache:                    cache.New(),
		plumber:                  models.NewAMQPPlumber(),
		pendingMetricDataRequest: make(map[string]models.RTSMetricConfig),
		pulling:                  make(map[int32]*ContainerPulling),
	}
}

func (s *RTS) Run() {
	s.log.Info("Starting listeners...")

	go t.HandleAPINotifications(s.amqph, &t.NotificationHandler{
		OnContainerUpdated:  s.onContainerUpdated,
		OnContainerDeleted:  s.onContainerDeleted,
		OnMetricUpdated:     s.onMetricUpdated,
		OnMetricDeleted:     s.onMetricDeleted,
		OnDataPolicyDeleted: s.onDataPolicyDeleted,
		OnError: func(err error) {
			s.log.Error("Error handling API notifications", logger.ErrField(err))
		},
	})
	go s.metricDataRequestListener()
	go s.metricDataListener()
	go s.globalMetricDataListener()
	go s.metricsDataListener()

	s.log.Info("Service is ready!")
	<-s.Done()
}

// Close connections.
func (s *RTS) Close() error {
	s.log.Close()
	s.amqp.Close()
	s.pg.Close()
	s.cache.Close()
	s.DispatchDone(nil)
	return nil
}
