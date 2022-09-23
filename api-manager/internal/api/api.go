package api

import (
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

// Create a new API instance
func New(pgConn *db.PgConn) (*API, error) {
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
