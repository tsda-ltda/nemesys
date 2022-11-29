package api

import (
	"context"
	"fmt"
	stdlog "log"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/auth"
	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/amqph"
	"github.com/fernandotsda/nemesys/shared/cache"
	"github.com/fernandotsda/nemesys/shared/env"
	"github.com/fernandotsda/nemesys/shared/influxdb"
	"github.com/fernandotsda/nemesys/shared/initdb"
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
}

func New() service.Service {
	amqpConn, err := amqp.Dial()
	if err != nil {
		stdlog.Panicf("Fail to dial with amqp server, err: %s", err)
		return nil
	}

	log, err := logger.New(amqpConn, logger.Config{
		Service:        "api-manager",
		ConsoleLevel:   logger.ParseLevelEnv(env.LogConsoleLevelAPIManager),
		BroadcastLevel: logger.ParseLevelEnv(env.LogBroadcastLevelAPIManager),
	})
	if err != nil {
		stdlog.Panicf("Fail to create logger, err: %s", err)
		return nil
	}
	log.Info("Connected to amqp server")

	initialized, err := initdb.PG()
	if err != nil {
		log.Error("Fail to inicialize database", logger.ErrField(err))
		return nil
	}
	if initialized {
		log.Info("Database inicialized with success")
	}

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
	amqph := amqph.New(amqpConn, log)
	api := &API{
		Tools:            service.NewTools(),
		PG:               pg.New(),
		Influx:           &influxClient,
		Router:           r,
		Auth:             auth,
		Validate:         validate,
		Log:              log,
		Cache:            cache.New(),
		Amqph:            amqph,
		UserPWBcryptCost: bcryptCost,
	}
	if !initialized {
		return api
	}

	err = CreateDefaultUser(context.Background(), api)
	if err != nil {
		log.Panic("Fail to create dafault user", logger.ErrField(err))
		return nil
	}
	log.Info("Default master user created")
	return api
}

func (api *API) Run() {
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
