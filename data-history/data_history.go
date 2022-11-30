package dhs

import (
	stdlog "log"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/amqph"
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

func New(serviceNumber service.NumberType) service.Service {
	// connect to amqp server
	amqpConn, err := amqp.Dial()
	if err != nil {
		stdlog.Panicf("Fail to dial with amqp server, err: %s", err)
		return nil
	}

	// create logger
	log, err := logger.New(amqpConn, logger.Config{
		Service:        "dhs",
		ConsoleLevel:   logger.ParseLevelEnv(env.LogConsoleLevelDHS),
		BroadcastLevel: logger.ParseLevelEnv(env.LogBroadcastLevelDHS),
	})
	if err != nil {
		stdlog.Panicf("Fail to create logger, err: %s", err)
		return nil
	}
	log.Info("Connected to amqp server")

	// connect influxdb
	influxClient, err := influxdb.Connect()
	if err != nil {
		log.Panic("Fail to connect to influxdb", logger.ErrField(err))
		return nil
	}
	log.Info("Connected to influxdb")

	tools := service.NewTools(service.RTS, serviceNumber)
	amqph := amqph.New(amqpConn, log, tools.ServiceIdent)
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
	go d.notificationListener()

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
