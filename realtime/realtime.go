package rts

import (
	stdlog "log"
	"sync"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/amqph"
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
	return &RTS{
		Tools:                    tools,
		log:                      log,
		pg:                       pg.New(),
		amqp:                     amqpConn,
		amqph:                    amqph.New(amqpConn, log, tools.ServiceIdent),
		cache:                    cache.New(),
		plumber:                  models.NewAMQPPlumber(),
		pendingMetricDataRequest: make(map[string]models.RTSMetricConfig),
		pulling:                  make(map[int32]*ContainerPulling),
	}
}

func (s *RTS) Run() {
	s.log.Info("Starting listeners...")
	go s.notificationListener()      // listen to notification
	go s.metricDataRequestListener() // listen to data requests
	go s.metricDataListener()        // listen to new data
	go s.metricsDataListener()       // listen to new data

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
