package rts

import (
	"context"
	"log"

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

	// pendingDataMap is a map of pending data requests.
	pendingDataMap map[string]models.RTSMetricInfo

	// getDataCh is the channel to request data.
	getDataCh chan models.AMQPCorrelated[[]byte]

	// publisherDataCh is the channel to publish data.
	publisherDataCh chan amqp091.Publishing

	// publisherDataCh is the channel to publish data.
	getDataSNMPCh chan amqp091.Publishing

	// stopDataRequestPublisher is the channel to stop the DataRequestPubliser
	stopDataRequestPublisher chan any

	// stopDataRequestListener is the channel to stop the DataRequestListener
	stopDataRequestListener chan any

	// stopDataListener is the channel to stop the DataListener
	stopDataListener chan any

	// stopDataPublisher is the channel to stop the DataPublisher
	stopDataPublisher chan any

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
	l.Info("connected to postgresql")

	return &RTS{
		pgConn:                   pg,
		amqp:                     amqpConn,
		Log:                      l,
		plumber:                  models.NewAMQPPlumber(),
		pendingDataMap:           make(map[string]models.RTSMetricInfo),
		publisherDataCh:          make(chan amqp091.Publishing),
		getDataCh:                make(chan models.AMQPCorrelated[[]byte]),
		getDataSNMPCh:            make(chan amqp091.Publishing),
		closed:                   make(chan any),
		stopDataRequestPublisher: make(chan any),
		stopDataListener:         make(chan any),
		stopDataRequestListener:  make(chan any),
		stopDataPublisher:        make(chan any),
		cache:                    cache.New(),
	}
}

func (s *RTS) Run() {
	go s.DataRequestListener()  // listen to data requests
	go s.DataPublisher()        // publish data responses
	go s.DataListener()         // listen to new data
	go s.DataRequestPublisher() // publish data requests fro translators

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

	s.stopDataListener <- nil
	s.stopDataPublisher <- nil
	s.stopDataRequestListener <- nil
	s.stopDataRequestPublisher <- nil

	s.closed <- nil
}
