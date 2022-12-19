package alarm

import (
	stdlog "log"
	"net/smtp"
	"strconv"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/amqph"
	t "github.com/fernandotsda/nemesys/shared/amqph/tools"
	"github.com/fernandotsda/nemesys/shared/cache"
	"github.com/fernandotsda/nemesys/shared/env"
	"github.com/fernandotsda/nemesys/shared/influxdb"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/pg"
	"github.com/fernandotsda/nemesys/shared/service"
	"github.com/rabbitmq/amqp091-go"
)

type Alarm struct {
	service.Tools

	// pg is the postgres handler.
	pg *pg.PG
	// cache is the cache handler.
	cache *cache.Cache
	// amqpConn is the amqp connection.
	amqpConn *amqp091.Connection
	// log is the logger.
	log *logger.Logger
	// amqph is the amqp handler.
	amqph *amqph.Amqph
	// smtpAuth is the smtp plain auth.
	smtpAuth smtp.Auth
	// influxdb is the influxdb client.
	influxdb *influxdb.Client
}

func New(serviceNumber int) service.Service {
	amqpConn, err := amqp.Dial()
	if err != nil {
		stdlog.Fatalf("Fail to dial with amqp server, err: %s", err)
		return nil
	}

	log, err := logger.New(amqpConn, logger.Config{
		Service:        service.GetServiceName(service.Alarm),
		ConsoleLevel:   logger.ParseLevelEnv(env.LogConsoleLevelAlarmService),
		BroadcastLevel: logger.ParseLevelEnv(env.LogBroadcastLevelAlarmService),
	})
	if err != nil {
		stdlog.Fatalf("Fail to create logger, err: %s", err)
		return nil
	}
	log.Info("Connected to amqp server")

	influxdb, err := influxdb.Connect()
	if err != nil {
		log.Fatal("Fail to connect to influxdb", logger.ErrField(err))
		return nil
	}
	log.Info("Connected to influxdb")

	tools := service.NewTools(service.Alarm, serviceNumber)

	publishers, err := strconv.Atoi(env.AlarmServiceAMQPPublishers)
	if err != nil {
		log.Fatal("Fail to parse env.AlarmServiceAMQPPublishers", logger.ErrField(err))
		return nil
	}

	amqph := amqph.New(amqph.Config{
		Log:        log,
		Conn:       amqpConn,
		Publishers: publishers,
	})
	go t.ServicePing(amqph, tools.ServiceIdent)

	cache, err := cache.New()
	if err != nil {
		log.Fatal("Fail to connect to cache (redis)", logger.ErrField(err))
		return nil
	}

	return &Alarm{
		pg:       pg.New(),
		cache:    cache,
		amqpConn: amqpConn,
		log:      log,
		amqph:    amqph,
		Tools:    tools,
		influxdb: &influxdb,
		smtpAuth: smtp.PlainAuth("", env.MetricAlarmEmailSender, env.MetricAlarmEmailSenderPassword, env.MetricAlarmEmailSenderHost),
	}
}

func (a *Alarm) Run() {
	a.log.Info("Starting listeners...")
	go a.listenCheckMetricsAlarm()
	go a.listenCheckMetricAlarm()
	go a.listenMetricsAlarmed()
	go a.listenMetricAlarmed()

	a.log.Info("Service is ready!")
	<-a.Done()
}

func (a *Alarm) Close() error {
	a.DispatchDone(nil)

	a.amqpConn.Close()
	a.amqph.Close()
	a.cache.Close()
	a.log.Close()
	a.pg.Close()
	a.influxdb.Close()

	return nil
}
