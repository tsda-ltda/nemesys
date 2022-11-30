package snmp

import (
	stdlog "log"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/amqph"
	"github.com/fernandotsda/nemesys/shared/cache"
	"github.com/fernandotsda/nemesys/shared/env"
	"github.com/fernandotsda/nemesys/shared/evaluator"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/pg"
	"github.com/fernandotsda/nemesys/shared/service"
	"github.com/rabbitmq/amqp091-go"
)

type SNMP struct {
	service.Tools
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
	// stopGetListener is the channel to stop the getListener
	stopGetListener chan any
	// stopDataListener is the channel to stop the dataPublisher
	stopDataPublisher chan any
}

func New(serviceNumber service.NumberType) service.Service {
	amqpConn, err := amqp.Dial()
	if err != nil {
		stdlog.Panicf("Fail to dial with amqp server, err: %s", err.Error())
		return nil
	}

	log, err := logger.New(amqpConn, logger.Config{
		Service:        "snmp",
		ConsoleLevel:   logger.ParseLevelEnv(env.LogConsoleLevelSNMP),
		BroadcastLevel: logger.ParseLevelEnv(env.LogBroadcastLevelSNMP),
	})
	if err != nil {
		stdlog.Panicf("Fail to create logger, err: %s", err.Error())
		return nil
	}
	log.Info("Connected to amqp server")

	pg := pg.New()
	tools := service.NewTools(service.RTS, serviceNumber)
	return &SNMP{
		Tools:             tools,
		amqph:             amqph.New(amqpConn, log, tools.ServiceIdent),
		amqpConn:          amqpConn,
		pg:                pg,
		log:               log,
		evaluator:         evaluator.New(pg),
		cache:             cache.New(),
		stopGetListener:   make(chan any),
		stopDataPublisher: make(chan any),
	}
}

func (s *SNMP) Run() {
	s.log.Info("Starting listeners...")
	go s.getMetricListener()  // listen to metric data requests
	go s.getMetricsListener() // listen to metrics data requests

	s.log.Info("Service is ready!")
	err := <-s.Done()
	if err != nil {
		s.log.Error("Service stopped with error", logger.ErrField(err))
		return
	}
	s.log.Info("Service stopped gracefully")
}

// Close all connections.
func (s *SNMP) Close() error {
	s.stopDataPublisher <- nil
	s.stopGetListener <- nil
	s.DispatchDone(nil)
	return nil
}
