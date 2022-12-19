package manager

import (
	stdlog "log"
	"strconv"
	"time"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/amqph"
	"github.com/fernandotsda/nemesys/shared/env"
	"github.com/fernandotsda/nemesys/shared/influxdb"
	"github.com/fernandotsda/nemesys/shared/initdb"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/service"
	"github.com/rabbitmq/amqp091-go"
)

type ServiceManager struct {
	service.Tools
	// amqpConn is the amqp connection.
	amqpConn *amqp091.Connection
	// amqph is the amqp handler
	amqph *amqph.Amqph
	// influxClient is the influxdb client.
	influxClient *influxdb.Client
	// log is the log.
	log *logger.Logger
	// services is all services registered.
	services []service.ServiceStatus
	// pingPlumber is the ping plumber.
	pingPlumber models.AMQPPlumber
	// pingInterval is the ping interval.
	pingInterval time.Duration
}

func Start() {
	loadEnv()

	amqpConn, err := amqp.Dial()
	if err != nil {
		stdlog.Fatalf("Fail to dial with amqp server, err: %s", err)
		return
	}

	log, err := logger.New(amqpConn, logger.Config{
		Service:        "service-manager",
		ConsoleLevel:   logger.ParseLevelEnv(env.LogConsoleLevelServiceManager),
		BroadcastLevel: logger.ParseLevelEnv(env.LogBroadcastLevelServiceManager),
	})
	if err != nil {
		stdlog.Fatalf("Fail to create logger, err: %s", err)
		return
	}
	log.Info("Connected to amqp server")

	influxClient, err := influxdb.Connect()
	if err != nil {
		log.Fatal("Fail to connect to influxdb", logger.ErrField(err))
		return
	}
	log.Info("Connected to influxdb")

	created, err := influxClient.CreateLogsBucket()
	if err != nil {
		log.Fatal("Fail to create logs bucket", logger.ErrField(err))
		return
	}
	if created {
		log.Info("Logs bucket created with success")
	}

	created, err = influxClient.CreateRequestsCountBucket()
	if err != nil {
		log.Fatal("Fail to create requets count bucket", logger.ErrField(err))
		return
	}
	if created {
		log.Info("Requests count bucket created with success")
	}

	created, err = influxClient.CreateAlarmHistoryBucket()
	if err != nil {
		log.Fatal("Fail to create alarm history bucket", logger.ErrField(err))
		return
	}
	if created {
		log.Info("ALarm history bucket created with success")
	}

	amqph := amqph.New(amqph.Config{
		Log:        log,
		Conn:       amqpConn,
		Publishers: 1,
	})

	s := ServiceManager{
		amqpConn:     amqpConn,
		influxClient: &influxClient,
		log:          log,
		services:     make([]service.ServiceStatus, 0),
		Tools:        service.NewTools(service.ServiceManager, 1),
		amqph:        amqph,
		pingPlumber:  *models.NewAMQPPlumber(),
	}

	initialized, err := initdb.PG()
	if err != nil {
		log.Fatal("Fail to inicialize postgres", logger.ErrField(err))
		return
	}
	if initialized {
		log.Info("Postgres inicialized with success")
	} else {
		log.Info("Database inicialization skipped")
	}

	interval, err := strconv.ParseInt(env.ServiceManagerPingInterval, 0, 64)
	if err != nil {
		s.log.Fatal("Fail to parse env.ServiceManagerPingInterval", logger.ErrField(err))
		return
	}
	s.pingInterval = time.Duration(interval) * time.Second

	log.Info("Starting listeners...")

	go s.registryListener()
	go s.pingHandler()
	go s.logListener()
	go s.pongHandler()

	log.Info("Service is ready!")
	<-s.Done()
}

func loadEnv() {
	err := env.LoadEnvFile()
	if err != nil {
		stdlog.Println("Fail to load env file")
	}
	env.Init()
}
