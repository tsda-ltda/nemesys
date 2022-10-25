package snmp

import (
	"log"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/db"
	"github.com/fernandotsda/nemesys/shared/env"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/rabbitmq/amqp091-go"
)

type SNMPService struct {
	// Log is the logger handler.
	Log *logger.Logger

	// amqp is the amqp connection.
	amqp *amqp091.Connection

	// pgConn is the postgresql connection.
	pgConn *db.PgConn

	// conns is a cache map of container id and snmp agent configuration and connection.
	conns map[int]*Conn

	// metrics is a cache map of metric id and metric.
	metrics map[int]*Metric

	// Done is the channel to quit.
	Done chan any

	// singleDataReq is the channel for new data requests.
	singleDataReq chan models.AMQPCorrelated[SingleDataReq]
}

// New returns a configurated SNMPService instance.
func New() *SNMPService {
	// connect to amqp server
	conn, err := amqp.Dial()
	if err != nil {
		log.Fatalf("fail to connect to amqp server, err: %s", err)
	}

	// create _logger
	_logger, err := logger.New(conn, logger.Config{
		Service:        "snmp",
		ConsoleLevel:   logger.ParseLevelEnv(env.LogConsoleLevelSNMP),
		BroadcastLevel: logger.ParseLevelEnv(env.LogBroadcastLevelSNMP),
	})
	if err != nil {
		log.Fatalf("fail to create logger, err: %s", err)
	}

	// connect to postgresql
	pgConn, err := db.ConnectToPG()
	if err != nil {
		_logger.Fatal("fail to connect to posgresql", logger.ErrField(err))
	}

	return &SNMPService{
		conns:         make(map[int]*Conn, 100),
		metrics:       make(map[int]*Metric),
		singleDataReq: make(chan models.AMQPCorrelated[SingleDataReq]),
		Done:          make(chan any),
		amqp:          conn,
		pgConn:        pgConn,
		Log:           _logger,
	}
}

// Run sets up all receivers and producers.
func (s *SNMPService) Run() {
	go s.dataProducer()
	go s.getListener()
	s.Log.Info("service running with success")
}

// Close all connections.
func (s *SNMPService) Close() {
	for _, c := range s.conns {
		if c != nil {
			c.Close()
		}
	}
	close(s.Done)
	s.Log.Info("service closed")
}
