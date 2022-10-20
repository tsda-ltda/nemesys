package snmp

import (
	"fmt"
	"log"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/env"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/rabbitmq/amqp091-go"
)

type SNMPService struct {
	Log     *logger.Logger
	amqp    *amqp091.Connection
	conns   map[string]*Conn
	Done    chan any
	dataReq chan DataReq
}

// New returns a configurated SNMPService instance.
func New() *SNMPService {
	// connect to amqp server
	conn, err := amqp.Dial()
	if err != nil {
		log.Fatalf("fail to connect to amqp server, err: %s", err)
	}

	// create logger
	logger, err := logger.New(conn, logger.Config{
		Service:        "snmp",
		ConsoleLevel:   logger.ParseLevelEnv(env.LogConsoleLevelSNMP),
		BroadcastLevel: logger.ParseLevelEnv(env.LogBroadcastLevelSNMP),
	})
	if err != nil {
		log.Fatalf("fail to create logger, err: %s", err)
	}

	return &SNMPService{
		conns:   make(map[string]*Conn, 100),
		dataReq: make(chan DataReq),
		amqp:    conn,
		Log:     logger,
	}
}

// connKey returns a key string for SNMPService.conns.
func connKey(target string, port uint16) string {
	return target + ":" + fmt.Sprint(port)
}

// Run sets up all receivers and producers.
func (s *SNMPService) Run() {
	go s.dataProducer()
	go s.getReceiver()
	go s.registerConnReceiver()
	s.Log.Info("service running with success")
}

// GetConn get a SNMP agent connection.
func (s *SNMPService) GetConn(target string, port uint16) *Conn {
	return s.conns[connKey(target, port)]
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
