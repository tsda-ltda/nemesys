package snmp

import (
	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/amqph"
	"github.com/fernandotsda/nemesys/shared/cache"
	"github.com/fernandotsda/nemesys/shared/env"
	"github.com/fernandotsda/nemesys/shared/evaluator"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/pg"
	"github.com/rabbitmq/amqp091-go"
)

type SNMPService struct {
	// log is the logger handler.
	log *logger.Logger
	// amqpConn is the amqp connection.
	amqpConn *amqp091.Connection
	// amqph is the amqp handler for common tasks.
	amqph *amqph.Amqph
	// pg is the postgresql handler.
	pg *pg.PG
	// evaluator is the metric evaluator
	evaluator *evaluator.Evaluator
	// cache is the cache handler.
	cache *cache.Cache
	// closed is the channel to quit.
	closed chan any
	// stopGetListener is the channel to stop the getListener
	stopGetListener chan any
	// stopDataListener is the channel to stop the dataPublisher
	stopDataPublisher chan any
}

// New returns a configurated SNMPService instance.
func New() *SNMPService {
	// connect to amqp server
	amqpConn, err := amqp.Dial()
	if err != nil {
		panic("fail to connect to amqp server, err: " + err.Error())
	}

	// create logger
	log, err := logger.New(amqpConn, logger.Config{
		Service:        "snmp",
		ConsoleLevel:   logger.ParseLevelEnv(env.LogConsoleLevelSNMP),
		BroadcastLevel: logger.ParseLevelEnv(env.LogBroadcastLevelSNMP),
	})
	if err != nil {
		panic("fail to create logger, err: " + err.Error())
	}

	pg := pg.New()
	return &SNMPService{
		amqph:             amqph.New(amqpConn, log),
		amqpConn:          amqpConn,
		pg:                pg,
		log:               log,
		evaluator:         evaluator.New(pg),
		cache:             cache.New(),
		stopGetListener:   make(chan any),
		stopDataPublisher: make(chan any),
	}
}

// Run sets up all receivers and producers.
func (s *SNMPService) Run() {
	s.log.Info("starting listeners...")
	go s.getMetricListener()  // listen to metric data requests
	go s.getMetricsListener() // listen to metrics data requests

	s.log.Info("service is ready!")
	<-s.closed
}

// Close all connections.
func (s *SNMPService) Close() {
	s.stopDataPublisher <- nil
	s.stopGetListener <- nil

	s.closed <- nil
	s.log.Info("service closed")
}
