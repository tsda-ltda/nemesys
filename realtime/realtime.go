package rts

import (
	"context"
	"log"
	"sync"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/amqph"
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
	// amqph is the amqp handler for common taks.
	amqph *amqph.Amqph
	// Log is the logger handler.
	Log *logger.Logger
	// plumber is the custom amqp message plumber
	// for deal with request responses cases.
	plumber *models.AMQPPlumber
	// muStartPulling is the mutex to add metrics on pulling.
	muStartPulling sync.Mutex
	// pulling is the map of containers pulling.
	pulling map[int32]*ContainerPulling
	// pendingMetricData is a map of pending data requests.
	pendingMetricData map[string]models.RTSMetricConfig
	// metricDataRequestCh is the channel to request metric data.
	metricDataRequestCh chan models.AMQPCorrelated[[]byte]
	// metricDataRequestCh is the channel to request metric data.
	metricsDataRequestCh chan models.AMQPCorrelated[[]byte]
	// metricDataPublisherCh is the channel to publish metric data.
	metricDataPublisherCh chan amqp091.Publishing
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

	// create amqp handler
	amqph := amqph.New(amqpConn, l)

	return &RTS{
		Log:                   l,
		pgConn:                pg,
		amqp:                  amqpConn,
		amqph:                 amqph,
		cache:                 cache.New(),
		plumber:               models.NewAMQPPlumber(),
		pendingMetricData:     make(map[string]models.RTSMetricConfig),
		pulling:               make(map[int32]*ContainerPulling),
		metricDataPublisherCh: make(chan amqp091.Publishing),
		metricDataRequestCh:   make(chan models.AMQPCorrelated[[]byte]),
		metricsDataRequestCh:  make(chan models.AMQPCorrelated[[]byte]),
		closed:                make(chan any),
	}
}

func (s *RTS) Run() {
	s.setupNotificationHandler()

	s.Log.Info("starting listeners...")
	go s.amqph.ContainerListener()   // listen to container notifications
	go s.amqph.MetricListener()      // listen to metric notification
	go s.MetricDataRequestListener() // listen to data requests
	go s.MetricDataListener()        // listen to new data
	go s.MetricsDataListener()       // listen to new data

	s.Log.Info("starting publishers...")
	go s.MetricDataRequestPublisher()  // publish metric data requests fro translators
	go s.MetricsDataRequestPublisher() // publish metrics data requests fro translators
	go s.MetricDataPublisher()         // publish data responses

	s.Log.Info("service is ready!")
	<-s.closed
}

func (s *RTS) setupNotificationHandler() {
	// close container pulling on update
	s.amqph.Notifications.Subscribe(amqph.ContainerUpdated, "", func(d any) {
		n := d.(amqph.ContainerNotification)
		c, ok := s.pulling[n.Base.Id]
		if !ok {
			return
		}
		c.Close()
	})

	// close container pulling on delete
	s.amqph.Notifications.Subscribe(amqph.ContainerDeleted, "", func(d any) {
		id := d.(int32)
		c, ok := s.pulling[id]
		if !ok {
			return
		}
		c.Close()
	})

	// update metric type and RTS info on update
	s.amqph.Notifications.Subscribe(amqph.MetricUpdated, "", func(d any) {
		n := d.(amqph.MetricNotification)
		c, ok := s.pulling[n.Base.ContainerId]
		if !ok {
			return
		}
		m, ok := c.Metrics[n.Base.Id]
		if !ok {
			return
		}
		m.Type = n.Base.Type
		m.RTSMetricConfig = models.RTSMetricConfig{
			PullingTimes:  n.Base.RTSPullingTimes,
			CacheDuration: n.Base.RTSCacheDuration,
		}
	})

	// stop metric pulling on metric delete
	s.amqph.Notifications.Subscribe(amqph.MetricDeleted, "", func(d any) {
		mp := d.(models.MetricPairId)
		cp, ok := s.pulling[mp.ContainerId]
		if !ok {
			return
		}
		m, ok := cp.Metrics[mp.Id]
		if !ok {
			return
		}
		m.Stop()
	})
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
	s.closed <- nil
}
