package rts

import (
	"context"
	"log"
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
	// pgConn is the postgresql connection.
	pgConn *pg.Conn
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
	// pendingMetricData is a map of pending data requests.
	pendingMetricData map[string]models.RTSMetricConfig
	// closed is the filled when rts is closed.
	closed chan any
}

// New returns a configurated RTS. Will kill if anything goes wrong.
func New() *RTS {
	// connect to amqp server
	amqpConn, err := amqp.Dial()
	if err != nil {
		log.Panicf("fail to connect to amqp server, err: %s", err)
		return nil
	}

	// create logger
	l, err := logger.New(
		amqpConn,
		logger.Config{
			Service:        "rts",
			ConsoleLevel:   logger.ParseLevelEnv(env.LogConsoleLevelRTS),
			BroadcastLevel: logger.ParseLevelEnv(env.LogBroadcastLevelRTS),
		},
	)
	if err != nil {
		log.Panicf("fail to create logger, err: %s", err)
		return nil
	}

	// connect to postgres
	pg, err := pg.Connect()
	if err != nil {
		l.Panic("fail to connect postgres", logger.ErrField(err))
	}
	l.Info("connected to postgres ")

	// create amqp handler
	amqph := amqph.New(amqpConn, l)

	return &RTS{
		log:               l,
		pgConn:            pg,
		amqp:              amqpConn,
		amqph:             amqph,
		cache:             cache.New(),
		plumber:           models.NewAMQPPlumber(),
		pendingMetricData: make(map[string]models.RTSMetricConfig),
		pulling:           make(map[int32]*ContainerPulling),
		closed:            make(chan any),
	}
}

func (s *RTS) Run() {

	s.log.Info("starting listeners...")
	go s.containerListener()         // listen to container changes
	go s.metricListener()            // listen to metric changes
	go s.metricDataRequestListener() // listen to data requests
	go s.metricDataListener()        // listen to new data
	go s.metricsDataListener()       // listen to new data

	s.log.Info("service is ready!")
	<-s.closed
}

// Close connections.
func (s *RTS) Close() {
	// Close logger
	s.log.Close()
	// Close amqp connection
	s.amqp.Close()
	// close postgresql
	s.pgConn.Close(context.Background())
	// close Redis client
	s.cache.Close()
	s.closed <- nil
}
