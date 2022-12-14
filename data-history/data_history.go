package dhs

import (
	stdlog "log"
	"strconv"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/amqph"
	t "github.com/fernandotsda/nemesys/shared/amqph/tools"
	"github.com/fernandotsda/nemesys/shared/env"
	"github.com/fernandotsda/nemesys/shared/influxdb"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/pg"
	"github.com/fernandotsda/nemesys/shared/service"
	"github.com/rabbitmq/amqp091-go"
)

type DHS struct {
	service.Tools
	// influxClient is the influx client.
	influxClient *influxdb.Client
	// pg is the postgres handler.
	pg *pg.PG
	// amqpConn is the amqp connection.
	amqpConn *amqp091.Connection
	// amqph is the amqp handler.
	amqph *amqph.Amqph
	// log is the internal logger.
	log *logger.Logger
	// containersPulling is a map of containersPulling.
	containersPulling map[string]*containerPulling
	// metricsContainerMap is a map of metricId and containerPulling key.
	metricsContainerMap map[int64]string
	// flexsLegacyPulling is a map of flexsLegacyPulling
	flexsLegacyPulling map[int32]*flexLegacyPulling
	// getFlexLegacyDatalogCh is the channel to get datalogs.
	getFlexLegacyDatalogCh chan int32
	// flexLegacyDatalogWorkers is the flex legacy datalog workers.
	flexLegacyDatalogWorkers []*flexLegacyDatalogWorker
	// IsReady is the service ready state.
	IsReady bool
}

func New(serviceNumber int) service.Service {
	tools := service.NewTools(service.DHS, serviceNumber)
	amqpConn, err := amqp.Dial()
	if err != nil {
		stdlog.Panicf("Fail to dial with amqp server, err: %s", err)
		return nil
	}

	log, err := logger.New(amqpConn, logger.Config{
		Service:        tools.ServiceIdent,
		ConsoleLevel:   logger.ParseLevelEnv(env.LogConsoleLevelDHS),
		BroadcastLevel: logger.ParseLevelEnv(env.LogBroadcastLevelDHS),
	})
	if err != nil {
		stdlog.Panicf("Fail to create logger, err: %s", err)
		return nil
	}
	log.Info("Connected to amqp server")

	influxClient, err := influxdb.Connect()
	if err != nil {
		log.Panic("Fail to connect to influxdb", logger.ErrField(err))
		return nil
	}
	log.Info("Connected to influxdb")

	publishers, err := strconv.Atoi(env.DHSAMQPPublishers)
	if err != nil {
		log.Fatal("Fail to parse env.DHSAMQPPublishers", logger.ErrField(err))
		return nil
	}

	amqph := amqph.New(amqph.Config{
		Log:        log,
		Conn:       amqpConn,
		Publishers: publishers,
	})
	go t.ServicePing(amqph, tools.ServiceIdent)

	return &DHS{
		Tools:                    tools,
		influxClient:             &influxClient,
		pg:                       pg.New(),
		amqpConn:                 amqpConn,
		amqph:                    amqph,
		log:                      log,
		containersPulling:        make(map[string]*containerPulling),
		metricsContainerMap:      make(map[int64]string),
		flexsLegacyPulling:       make(map[int32]*flexLegacyPulling),
		getFlexLegacyDatalogCh:   make(chan int32),
		flexLegacyDatalogWorkers: make([]*flexLegacyDatalogWorker, 0),
		IsReady:                  false,
	}
}

func (d *DHS) Run() {
	d.createFlexLegacyWorkers()
	err := d.readDatabase()
	if err != nil {
		d.log.Panic("Fail to read database", logger.ErrField(err))
		return
	}
	d.log.Info("Starting listeners...")

	go d.metricsDataListener()
	go t.HandleAPINotifications(d.amqph, &t.NotificationHandler{
		ContainerCreatedQueue: amqp.QueueDHSContainerCreated,
		MetricCreatedQueue:    amqp.QueueDHSMetricCreated,
		OnContainerCreated:    d.onContainerCreated,
		OnContainerUpdated:    d.onContainerUpdated,
		OnContainerDeleted:    d.onContainerDeleted,
		OnMetricCreated:       d.onMetricCreated,
		OnMetricUpdated:       d.onMetricUpdated,
		OnMetricDeleted:       d.onMetricDeleted,
		OnDataPolicyDeleted:   d.onDataPolicyDeleted,

		OnError: func(err error) {
			d.log.Error("Error handling API notifications", logger.ErrField(err))
		},
	})

	d.IsReady = true
	d.log.Info("Service is ready!")

	err = <-d.Done()
	if err != nil {
		d.log.Error("Service stopped with error", logger.ErrField(err))
		return
	}
	d.log.Info("Service stopped gracefully")
}

func (d *DHS) Close() error {
	d.amqpConn.Close()
	d.influxClient.Close()
	d.pg.Close()
	d.DispatchDone(nil)
	return nil
}
