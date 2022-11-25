package rts

import (
	"sync"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/amqph"
	"github.com/fernandotsda/nemesys/shared/cache"
	"github.com/fernandotsda/nemesys/shared/env"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/pg"
	"github.com/rabbitmq/amqp091-go"
)

// Real Time Service
type RTS struct {
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
	// closed is the filled when rts is closed.
	closed chan any
}

// New returns a configurated RTS. Will kill if anything goes wrong.
func New() *RTS {
	amqpConn, err := amqp.Dial()
	if err != nil {
		panic("fail to connect to amqp server, err: " + err.Error())
	}

	log, err := logger.New(
		amqpConn,
		logger.Config{
			Service:        "rts",
			ConsoleLevel:   logger.ParseLevelEnv(env.LogConsoleLevelRTS),
			BroadcastLevel: logger.ParseLevelEnv(env.LogBroadcastLevelRTS),
		},
	)
	if err != nil {
		panic("fail to create logger, err: " + err.Error())
	}

	amqph := amqph.New(amqpConn, log)
	return &RTS{
		log:                      log,
		pg:                       pg.New(),
		amqp:                     amqpConn,
		amqph:                    amqph,
		cache:                    cache.New(),
		plumber:                  models.NewAMQPPlumber(),
		pendingMetricDataRequest: make(map[string]models.RTSMetricConfig),
		pulling:                  make(map[int32]*ContainerPulling),
		closed:                   make(chan any),
	}
}

func (s *RTS) Run() {
	s.log.Info("starting listeners...")
	go s.notificationListener()      // listen to notification
	go s.metricDataRequestListener() // listen to data requests
	go s.metricDataListener()        // listen to new data
	go s.metricsDataListener()       // listen to new data

	s.log.Info("service is ready!")
	<-s.closed
}

// Close connections.
func (s *RTS) Close() {
	s.log.Close()
	s.amqp.Close()
	s.pg.Close()
	s.cache.Close()
	s.closed <- nil
}
