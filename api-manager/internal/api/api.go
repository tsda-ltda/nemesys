package api

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/auth"
	"github.com/fernandotsda/nemesys/shared/amqph"
	"github.com/fernandotsda/nemesys/shared/cache"
	"github.com/fernandotsda/nemesys/shared/env"
	"github.com/fernandotsda/nemesys/shared/influxdb"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/pg"
	"github.com/fernandotsda/nemesys/shared/rdb"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/rabbitmq/amqp091-go"
)

type API struct {
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
	// Closed is filled when application is closed.
	Closed chan any
}

// Create a new API instance.
func New(conn *amqp091.Connection, log *logger.Logger) (*API, error) {
	// connect to redis auth
	rdbAuth, err := rdb.NewAuthClient()
	if err != nil {
		return nil, err
	}
	log.Info("connected to redis")

	// create authentication handler
	auth, err := auth.New(rdbAuth)
	if err != nil {
		return nil, err
	}

	// connect to influxdb
	influxClient, err := influxdb.Connect()
	if err != nil {
		return nil, err
	}
	log.Info("connected to influxdb")

	// parse bcrypt coast
	bcryptCost, err := strconv.Atoi(env.UserPWBcryptCost)
	if err != nil {
		return nil, errors.New("fail to parse env.UserPWBcryptCost, err: " + err.Error())
	}

	// create gin engine
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// create validator
	validate := validator.New()

	// create and run amqp handler
	amqph := amqph.New(conn, log)
	return &API{
		PG:               pg.New(),
		Influx:           &influxClient,
		Router:           r,
		Auth:             auth,
		Validate:         validate,
		Log:              log,
		Cache:            cache.New(),
		Amqph:            amqph,
		UserPWBcryptCost: bcryptCost,
		Closed:           make(chan any),
	}, nil
}

// Start listen and server.
func (api *API) Run() error {
	url := fmt.Sprintf("%s:%s", env.APIManagerHost, env.APIManagerPort)
	api.Log.Info("server listening to: " + url)
	return api.Router.Run(url)
}

// Close all api dependencies.
func (api *API) Close() {
	api.PG.Close()
	api.Auth.Close()
	api.Cache.Close()
	api.Amqph.Close()
}
