package api

import (
	"context"
	"fmt"
	stdlog "log"
	"strconv"
	"sync"
	"time"

	"github.com/fernandotsda/nemesys/api-manager/internal/auth"
	"github.com/fernandotsda/nemesys/api-manager/internal/counter"
	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/amqph"
	t "github.com/fernandotsda/nemesys/shared/amqph/tools"
	"github.com/fernandotsda/nemesys/shared/cache"
	"github.com/fernandotsda/nemesys/shared/env"
	"github.com/fernandotsda/nemesys/shared/influxdb"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/pg"
	"github.com/fernandotsda/nemesys/shared/rdb"
	"github.com/fernandotsda/nemesys/shared/service"
	"github.com/fernandotsda/nemesys/shared/trap"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/rabbitmq/amqp091-go"
)

type APIResponse struct {
	Data    any    `json:"data"`
	Message string `json:"message"`
}

type API struct {
	service.Tools
	// amqpConn is the amqp connection.
	amqpConn *amqp091.Connection
	// Amqph is the amqp handler.
	Amqph *amqph.Amqph
	// Postgresql handler.
	PG *pg.PG
	// Influx is the influxdb client.
	Influx *influxdb.Client
	// Cache is the cache handler.
	Cache *cache.Cache
	// Gin web fremework engine.
	Router *gin.Engine
	// Auth handler.
	Auth *auth.Auth
	// Validator.
	Validate *validator.Validate
	// User pw hash cost.
	UserPWBcryptCost int
	// Logger is the internal logger.
	Log *logger.Logger
	// Counter is the request counter.
	Counter *counter.Counter
	// servicesStatus are the current status of all registered
	// services by the service manager.
	servicesStatus []service.ServiceStatus
	// trapHandlers are the trap handlers.
	trapsListeners []*trap.Trap
	// trapHandlersMU is the mutex for create
	// and remove trap listeners.
	trapHandlersMU sync.Mutex
}

func New(serviceNumber int) service.Service {
	tools := service.NewTools(service.APIManager, serviceNumber)

	amqpConn, err := amqp.Dial()
	if err != nil {
		stdlog.Panicf("Fail to dial with amqp server, err: %s", err)
		return nil
	}

	log, err := logger.New(amqpConn, logger.Config{
		Service:        tools.ServiceIdent,
		ConsoleLevel:   logger.ParseLevelEnv(env.LogConsoleLevelAPIManager),
		BroadcastLevel: logger.ParseLevelEnv(env.LogBroadcastLevelAPIManager),
	})
	if err != nil {
		stdlog.Panicf("Fail to create logger, err: %s", err)
		return nil
	}
	log.Info("Connected to amqp server")

	rdbAuth, err := rdb.NewAuthClient()
	if err != nil {
		log.Panic("Fail to create auth client", logger.ErrField(err))
		return nil
	}
	log.Info("Connected to redis client (auth client)")

	auth, err := auth.New(rdbAuth)
	if err != nil {
		log.Panic("Fail to create auth handler", logger.ErrField(err))
		return nil
	}

	influxClient, err := influxdb.Connect()
	if err != nil {
		log.Panic("Fail to connect to influxdb", logger.ErrField(err))
		return nil
	}
	log.Info("Connected to influxdb")

	bcryptCost, err := strconv.Atoi(env.UserPWBcryptCost)
	if err != nil {
		log.Fatal("Fail to parse env.UserPWBcryptCost", logger.ErrField(err))
		return nil
	}

	validate := validator.New()
	pg := pg.New()

	publishers, err := strconv.Atoi(env.APIManagerAMQPPublishers)
	if err != nil {
		log.Fatal("Fail to parse env.APIManagerAMQPPublishers", logger.ErrField(err))
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

	api := &API{
		amqpConn:         amqpConn,
		Tools:            tools,
		PG:               pg,
		Influx:           &influxClient,
		Auth:             auth,
		Validate:         validate,
		Log:              log,
		Cache:            cache,
		Amqph:            amqph,
		UserPWBcryptCost: bcryptCost,
		Counter:          counter.New(&influxClient, pg, log, time.Second*10),
		servicesStatus:   []service.ServiceStatus{},
		trapsListeners:   []*trap.Trap{},
	}

	err = api.createDefaultUser(context.Background())
	if err != nil {
		log.Panic("Fail to create dafault user", logger.ErrField(err))
		return nil
	}
	return api
}

func (api *API) Run() {
	go api.servicesStatusListener()
	go api.startTrapListeners()

	url := fmt.Sprintf("%s:%s", env.APIManagerHost, env.APIManagerPort)
	api.Log.Info("Server listening to: " + url)

	err := api.Router.Run(url)
	if err != nil {
		api.Log.Error("Server stopped with error", logger.ErrField(err))
		return
	}
	api.Log.Info("Server stopped gracefully")
}

func (api *API) Close() error {
	api.PG.Close()
	api.Cache.Close()
	api.Amqph.Close()
	api.Auth.Close()
	api.DispatchDone(nil)
	return nil
}
