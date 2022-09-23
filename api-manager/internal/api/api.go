package api

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/fernandotsda/nemesys/shared/db"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type API struct {
	// Postgresql connection
	PgConn *db.PgConn

	// Gin web fremework engine
	Router *gin.Engine

	// Validator
	Validate *validator.Validate
}

// Create a new API instance.
func New() (*API, error) {
	// connect to postgresql
	pgConn, err := db.ConnectToPG(os.Getenv("PG_URL"))
	if err != nil {
		return nil, fmt.Errorf("fail to connect to pg, err: %s", err)
	}
	log.Println("connected to postgresql")

	// create gin engine
	r := gin.New()

	// create validator
	validate := validator.New()

	return &API{
		PgConn:   pgConn,
		Router:   r,
		Validate: validate,
	}, nil
}

// Start listen and server
func (api *API) Run(addr string) error {
	return api.Router.Run(addr)
}

// Close all api dependencies. It's safe to call Close
// on a already closed one
func (api *API) Close() {
	api.PgConn.Close(context.Background())
}
