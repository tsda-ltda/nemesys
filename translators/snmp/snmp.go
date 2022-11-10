package snmp

import (
	"log"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/amqph"
	"github.com/fernandotsda/nemesys/shared/db"
	"github.com/fernandotsda/nemesys/shared/env"
	"github.com/fernandotsda/nemesys/shared/evaluator"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/rabbitmq/amqp091-go"
)

type SNMPService struct {
	// Log is the logger handler.
	Log *logger.Logger
	// amqp is the amqp connection.
	amqp *amqp091.Connection
	// amqph is the amqp handler for common tasks.
	amqph *amqph.Amqph
	// pgConn is the postgresql connection.
	pgConn *db.PgConn
	// evaluator is the metric evaluator
	evaluator *evaluator.Evaluator
	// conns is a cache map of container id and snmp agent configuration and connection.
	conns map[int32]*Conn
	// metrics is a cache map of metric id and metric.
	metrics map[int64]*Metric
	// closed is the channel to quit.
	closed chan any
	// metricDataReq is the channel for new metric data requests.
	metricDataReq chan models.AMQPCorrelated[metricRequest]
	// metricsDataReq is the channel for new metrics data requests.
	metricsDataReq chan models.AMQPCorrelated[metricsRequest]
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
		log.Fatalf("fail to connect to amqp server, err: %s", err)
	}

	// create logger
	l, err := logger.New(conn, logger.Config{
		Service:        "snmp",
		ConsoleLevel:   logger.ParseLevelEnv(env.LogConsoleLevelSNMP),
		BroadcastLevel: logger.ParseLevelEnv(env.LogBroadcastLevelSNMP),
	})
	if err != nil {
		log.Fatalf("fail to create logger, err: %s", err)
	}

	// connect to postgres
	pgConn, err := db.ConnectToPG()
	if err != nil {
		l.Fatal("fail to connect to posgres", logger.ErrField(err))
	}

	return &SNMPService{
		amqph:             amqph.New(conn, l),
		amqp:              conn,
		pgConn:            pgConn,
		Log:               l,
		evaluator:         evaluator.New(pgConn),
		conns:             make(map[int32]*Conn, 100),
		metrics:           make(map[int64]*Metric),
		metricDataReq:     make(chan models.AMQPCorrelated[metricRequest]),
		metricsDataReq:    make(chan models.AMQPCorrelated[metricsRequest]),
		stopGetListener:   make(chan any),
		stopDataPublisher: make(chan any),
	}
}

// Run sets up all receivers and producers.
func (s *SNMPService) Run() {
	s.setupNotificationsHandler()

	s.Log.Info("starting listeners...")
	go s.amqph.MetricListener()    // listen to metrics updates
	go s.amqph.ContainerListener() // listen to containers updates
	go s.getMetricListener()       // listen to metric data requests
	go s.getMetricsListener()      // listen to metrics data requests

	s.Log.Info("starting publishers...")
	go s.metricDataPublisher()  // publish metric data
	go s.metricsDataPublisher() // publish metrics data

	s.Log.Info("service is ready!")
	<-s.closed
}

func (s *SNMPService) setupNotificationsHandler() {
	// close connection on container update
	s.amqph.Notifications.Subscribe(amqph.ContainerUpdated, "", func(d any) {
		n := d.(amqph.ContainerNotification)
		c, ok := s.conns[n.Base.Id]
		if !ok {
			return
		}
		c.Close()
	})

	// close connection on container delete
	s.amqph.Notifications.Subscribe(amqph.ContainerDeleted, "", func(d any) {
		id := d.(int32)
		c, ok := s.conns[id]
		if !ok {
			return
		}
		c.Close()
	})

	// close metric on update
	s.amqph.Notifications.Subscribe(amqph.MetricUpdated, "", func(d any) {
		n := d.(amqph.MetricNotification)
		m, ok := s.metrics[n.Base.Id]
		if !ok {
			return
		}
		m.Close()
	})

	// close metric on delete
	s.amqph.Notifications.Subscribe(amqph.MetricDeleted, "", func(d any) {
		n := d.(models.MetricPairId)
		m, ok := s.metrics[n.Id]
		if !ok {
			return
		}
		m.Close()
	})
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
	s.Log.Info("service closed")
}
