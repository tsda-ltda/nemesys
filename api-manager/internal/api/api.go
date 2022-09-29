package api

import (
	"context"
	"fmt"
	"log"

	"github.com/fernandotsda/nemesys/api-manager/internal/auth"
	_db "github.com/fernandotsda/nemesys/api-manager/internal/db"
	"github.com/fernandotsda/nemesys/shared/db"
	"github.com/fernandotsda/nemesys/shared/env"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type API struct {
	// Postgresql connection
	PgConn *db.PgConn

	// Gin web fremework engine
	Router *gin.Engine

	// Auth handler
	Auth *auth.Auth

	// Validator
	Validate *validator.Validate

	// User pw hash cost
	UserPWBcryptCost int
}

// Create a new API instance.
func New() (*API, error) {
	// connect to postgresql
	pgConn, err := _db.PGConnectAndInit()
	if err != nil {
		return nil, err
	}
	log.Println("connected to postgresql")

	// connect to redis auth
	rdbAuth, err := _db.RDBAuthConnectAndInit()
	if err != nil {
		return nil, err
	}
	log.Println("connected to redis")

	// create authentication handler
	auth, err := auth.New(rdbAuth)
	if err != nil {
		return nil, fmt.Errorf("fail to create auth handler, err: %s", err)
	}

	// create gin engine
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// create validator
	validate := validator.New()

	return &API{
		PgConn:   pgConn,
		Router:   r,
		Auth:     auth,
		Validate: validate,
	}, nil
}

// Start listen and server
func (api *API) Run() error {
	url := fmt.Sprintf("%s:%s", env.APIManagerHost, env.APIManagerPort)
	log.Printf("server listen to: %s", url)
	return api.Router.Run(url)
}

// Close all api dependencies. It's safe to call Close
// on a already closed one
func (api *API) Close() {
	api.PgConn.Close(context.Background())
	api.Auth.Close()
}
