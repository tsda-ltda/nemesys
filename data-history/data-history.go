package dhs

import (
	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/amqph"
	"github.com/fernandotsda/nemesys/shared/env"
	"github.com/fernandotsda/nemesys/shared/influxdb"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/pg"
	"github.com/rabbitmq/amqp091-go"
)

type DHS struct {
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
	// dataPullingGroups is a map of DataPullingGroup.
	dataPullingGroups map[string]*DataPullingGroup
	// metricsGroupMap is a map of metricId and DataPullingGroup key.
	metricsGroupsMap map[int64]string
	// close is the closed channel
	closed chan any
	// IsReady is the service ready state.
	IsReady bool
}

func New() (*DHS, error) {
	// connect to amqp server
	amqpConn, err := amqp.Dial()
	if err != nil {
		return nil, err
	}

	// create logger
	log, err := logger.New(amqpConn, logger.Config{
		Service:        "dhs",
		ConsoleLevel:   logger.ParseLevelEnv(env.LogConsoleLevelDHS),
		BroadcastLevel: logger.ParseLevelEnv(env.LogBroadcastLevelDHS),
	})
	if err != nil {
		return nil, err
	}

	// connect influxdb
	influxClient, err := influxdb.Connect()
	if err != nil {
		log.Error("fail to connect to influxdb", logger.ErrField(err))
		return nil, err
	}
	log.Info("connected to influxdb")

	// create amqph
	amqph := amqph.New(amqpConn, log)

	return &DHS{
		influxClient:      &influxClient,
		pg:                pg.New(),
		amqpConn:          amqpConn,
		amqph:             amqph,
		log:               log,
		dataPullingGroups: make(map[string]*DataPullingGroup),
		metricsGroupsMap:  make(map[int64]string),
	}, nil
}

func (d *DHS) Run() {
	err := d.readDatabase(100, 0)
	if err != nil {
		d.log.Panic("fail to read database", logger.ErrField(err))
	}
	d.log.Info("starting listeners...")
	go d.metricsDataListener()
	go d.containerListener()
	go d.metricListener()

	d.IsReady = true
	d.log.Info("service is ready!")

	<-d.closed
}

func (d *DHS) Close() {
	d.amqpConn.Close()
	d.influxClient.Close()
	d.pg.Close()
	d.closed <- nil
}
