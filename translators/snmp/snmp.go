package snmp

import (
	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/amqph"
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
	// pgConn is the postgresql connection.
	pgConn *pg.Conn
	// evaluator is the metric evaluator
	evaluator *evaluator.Evaluator
	// conns is a cache map of container id and snmp agent configuration and connection.
	conns map[int32]*ContainerConn
	// metrics is a cache map of metric id and metric.
	metrics map[int64]*Metric
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
	conn, err := amqp.Dial()
	if err != nil {
		panic("fail to connect to amqp server, err: " + err.Error())
	}

	// create logger
	log, err := logger.New(conn, logger.Config{
		Service:        "snmp",
		ConsoleLevel:   logger.ParseLevelEnv(env.LogConsoleLevelSNMP),
		BroadcastLevel: logger.ParseLevelEnv(env.LogBroadcastLevelSNMP),
	})
	if err != nil {
		panic("fail to create logger, err: " + err.Error())
	}

	// connect to postgres
	pgConn, err := pg.Connect()
	if err != nil {
		log.Fatal("fail to connect to posgres", logger.ErrField(err))
	}
	log.Info("connected to postgres")

	return &SNMPService{
		amqph:             amqph.New(conn, log),
		amqpConn:          conn,
		pgConn:            pgConn,
		log:               log,
		evaluator:         evaluator.New(pgConn),
		conns:             make(map[int32]*ContainerConn, 100),
		metrics:           make(map[int64]*Metric),
		stopGetListener:   make(chan any),
		stopDataPublisher: make(chan any),
	}
}

// Run sets up all receivers and producers.
func (s *SNMPService) Run() {
	s.log.Info("starting listeners...")
	go s.containerListener()  // listen to container changes
	go s.metricListener()     // listen to metric changes
	go s.getMetricListener()  // listen to metric data requests
	go s.getMetricsListener() // listen to metrics data requests

	s.log.Info("service is ready!")
	<-s.closed
}

// Close all connections.
func (s *SNMPService) Close() {
	for _, c := range s.conns {
		if c != nil {
			c.Close()
		}
	}

	s.stopDataPublisher <- nil
	s.stopGetListener <- nil

	s.closed <- nil
	s.log.Info("service closed")
}
