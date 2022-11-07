package rts

import (
	"context"
	"log"
	"sync"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/cache"
	"github.com/fernandotsda/nemesys/shared/db"
	"github.com/fernandotsda/nemesys/shared/env"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/rabbitmq/amqp091-go"
)

// Real Time Service
type RTS struct {
	// cache is the cache handler
	cache *cache.Cache

	// pgConn is the postgresql connection.
	pgConn *db.PgConn

	// amqp is the amqp connection.
	amqp *amqp091.Connection

	// Log is the logger handler.
	Log *logger.Logger

	// plumber is the custom amqp message plumber
	// for deal with request responses cases.
	plumber *models.AMQPPlumber

	// muPulling is the mutex to add metrics on pulling.
	muPulling sync.Mutex

	// pulling is the map of containers pulling.
	pulling map[int32]*ContainerPulling

	// pendingMetricData is a map of pending data requests.
	pendingMetricData map[string]models.RTSMetricInfo

	// pendingMetricData is a map of pending data requests.
	pendingMetricsData map[int32][]struct {
		Id int64
		models.RTSMetricInfo
	}

	// metricDataRequestCh is the channel to request metric data.
	metricDataRequestCh chan models.AMQPCorrelated[[]byte]

	// metricDataRequestCh is the channel to request metric data.
	metricsDataRequestCh chan models.AMQPCorrelated[[]byte]

	// metricDataPublisherCh is the channel to publish metric data.
	metricDataPublisherCh chan amqp091.Publishing

	// stopMetricDataRequestPublisher is the channel to stop the MetricDataRequestPublisher
	stopMetricDataRequestPublisher chan any

	// stopMetricsDataRequestPublisher is the channel to stop the MetricsDataRequestPublisher
	stopMetricsDataRequestPublisher chan any

	// stopMetricDataRequestListener is the channel to stop the DataRequestListener
	stopMetricDataRequestListener chan any

	// stopMetricDataListener is the channel to stop the DataListener
	stopMetricDataListener chan any

	// stopMetricDataListener is the channel to stop the DataListener
	stopMetricsDataListener chan any

	// stopMetricDataPublisher is the channel to stop the DataPublisher
	stopMetricDataPublisher chan any

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
	pg, err := db.ConnectToPG()
	if err != nil {
		l.Panic("fail to connect postgres", logger.ErrField(err))
	}
	l.Info("connected to postgres ")

	return &RTS{
		Log:     l,
		pgConn:  pg,
		amqp:    amqpConn,
		cache:   cache.New(),
		plumber: models.NewAMQPPlumber(),
		pendingMetricsData: make(map[int32][]struct {
			Id int64
			models.RTSMetricInfo
		}),
		pendingMetricData:               make(map[string]models.RTSMetricInfo),
		pulling:                         make(map[int32]*ContainerPulling),
		metricDataPublisherCh:           make(chan amqp091.Publishing),
		metricDataRequestCh:             make(chan models.AMQPCorrelated[[]byte]),
		metricsDataRequestCh:            make(chan models.AMQPCorrelated[[]byte]),
		stopMetricDataRequestPublisher:  make(chan any),
		stopMetricsDataRequestPublisher: make(chan any),
		stopMetricDataListener:          make(chan any),
		stopMetricsDataListener:         make(chan any),
		stopMetricDataRequestListener:   make(chan any),
		stopMetricDataPublisher:         make(chan any),
		closed:                          make(chan any),
	}
}

func (s *RTS) Run() {
	go s.MetricDataRequestListener()   // listen to data requests
	go s.MetricDataPublisher()         // publish data responses
	go s.MetricDataListener()          // listen to new data
	go s.MetricsDataListener()         // listen to new data
	go s.MetricDataRequestPublisher()  // publish metric data requests fro translators
	go s.MetricsDataRequestPublisher() // publish metrics data requests fro translators

	s.Log.Info("service started")
	<-s.closed
}

// Close connections.
func (s *RTS) Close() {
	// Close logger
	s.Log.Close()

	// Close amqp connection
	s.amqp.Close()

	// close postgresql
	s.pgConn.Close(context.Background())

	// close Redis client
	s.cache.Close()

	s.stopMetricDataListener <- nil
	s.stopMetricDataPublisher <- nil
	s.stopMetricDataRequestListener <- nil
	s.stopMetricDataRequestPublisher <- nil

	s.closed <- nil
}
