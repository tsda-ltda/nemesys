package router

import (
	"github.com/fernandotsda/nemesys/api-manager/internal/user"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
)

// Set all api routes
func Set(api *api.API) {
	// get api routes
	r := api.Router

	// user CRUD
	r.GET("/users", user.MGetHandler(api))
	r.POST("/users", user.CreateHandler(api))
	r.GET("/users/:id", user.GetHandler(api))
	r.PATCH("/users/:id", user.UpdateHandler(api))
	r.DELETE("/users/:id", user.DeleteHandler(api))
}
