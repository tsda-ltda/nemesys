package api

import (
	"context"
	"fmt"
	stdlog "log"
	"strconv"
	"time"

	"github.com/fernandotsda/nemesys/api-manager/internal/auth"
	"github.com/fernandotsda/nemesys/api-manager/internal/counter"
	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/amqph"
	"github.com/fernandotsda/nemesys/shared/cache"
	"github.com/fernandotsda/nemesys/shared/env"
	"github.com/fernandotsda/nemesys/shared/influxdb"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/pg"
	"github.com/fernandotsda/nemesys/shared/rdb"
	"github.com/fernandotsda/nemesys/shared/service"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type API struct {
	service.Tools

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

	gin.SetMode(gin.ReleaseMode)
	log.Info("Gin mode setted to 'release'")
	r := gin.New()

	validate := validator.New()
	pg := pg.New()
	api := &API{
		Tools:            tools,
		PG:               pg,
		Influx:           &influxClient,
		Router:           r,
		Auth:             auth,
		Validate:         validate,
		Log:              log,
		Cache:            cache.New(),
		Amqph:            amqph.New(amqpConn, log, tools.ServiceIdent),
		UserPWBcryptCost: bcryptCost,
		Counter:          counter.New(&influxClient, pg, log, time.Second*10),
		servicesStatus:   []service.ServiceStatus{},
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
