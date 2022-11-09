package api

import (
	"context"
	"fmt"

	"github.com/fernandotsda/nemesys/api-manager/internal/amqph"
	"github.com/fernandotsda/nemesys/api-manager/internal/auth"
	"github.com/fernandotsda/nemesys/shared/cache"
	"github.com/fernandotsda/nemesys/shared/db"
	"github.com/fernandotsda/nemesys/shared/env"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/rabbitmq/amqp091-go"
)

type API struct {
	// Amqp is the amqp connection.
	Amqp *amqp091.Connection

	// Amqph is the amqp handler.
	Amqph *amqph.Amqph

	// Postgresql connection.
	PgConn *db.PgConn

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
	// connect to postgresql
	pgConn, err := db.ConnectToPG()
	if err != nil {
		return nil, err
	}
	log.Info("connected to postgresql")

	// connect to redis auth
	rdbAuth, err := db.RDBAuthConnect()
	if err != nil {
		return nil, err
	}
	log.Info("connected to redis")

	// create authentication handler
	auth, err := auth.New(rdbAuth)
	if err != nil {
		return nil, err
	}

	// create gin engine
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// create validator
	validate := validator.New()

	return &API{
		Amqp:     conn,
		PgConn:   pgConn,
		Router:   r,
		Auth:     auth,
		Validate: validate,
		Log:      log,
		Cache:    cache.New(),
		Amqph:    amqph.New(conn, log),
	}, nil
}

// Start listen and server
func (api *API) Run() error {
	url := fmt.Sprintf("%s:%s", env.APIManagerHost, env.APIManagerPort)
	api.Log.Info("server listening to: " + url)
	return api.Router.Run(url)
}

// Close all api dependencies. It's safe to call Close
// on a already closed one
func (api *API) Close() {
	api.PgConn.Close(context.Background())
	api.Auth.Close()
	api.Cache.Close()
	api.Amqp.Close()
}
