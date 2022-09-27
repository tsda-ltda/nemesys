package router

import (
	"github.com/fernandotsda/nemesys/api-manager/internal/team"
	"github.com/fernandotsda/nemesys/api-manager/internal/user"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
)

// Set all api routes
func Set(api *api.API) {
	// get api routes
	r := api.Router

	// users CRUD
	r.GET("/users", user.MGetHandler(api))
	r.POST("/users", user.CreateHandler(api))
	r.GET("/users/:id", user.GetHandler(api))
	r.PATCH("/users/:id", user.UpdateHandler(api))
	r.DELETE("/users/:id", user.DeleteHandler(api))

	// teams CRUD
	r.GET("/teams", team.MGetHandler(api))
	r.POST("/teams", team.CreateHandler(api))
	r.PATCH("/teams/:ident", team.UpdateHandler(api))
	r.GET("/teams/:ident", team.GetHandler(api))
	r.DELETE("/teams/:ident", team.DeleteHandler(api))
	r.PATCH("/teams/:ident/users", team.UsersHandler(api))
}