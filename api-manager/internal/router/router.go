package router

import (
	"github.com/fernandotsda/nemesys/api-manager/internal/middleware"
	"github.com/fernandotsda/nemesys/api-manager/internal/roles"
	"github.com/fernandotsda/nemesys/api-manager/internal/team"
	"github.com/fernandotsda/nemesys/api-manager/internal/user"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
)

// Set all api routes
func Set(api *api.API) {
	// get api routes
	r := api.Router

	// login
	r.POST("/login", user.LoginHandler(api))

	// users
	r.GET("/users", middleware.Protect(api, roles.TeamsManager), user.MGetHandler(api))
	r.POST("/users", middleware.Protect(api, roles.Admin), user.CreateHandler(api))
	r.GET("/users/:id", middleware.ProtectUser(api, roles.Admin), user.GetHandler(api))
	r.PATCH("/users/:id", middleware.Protect(api, roles.Admin), user.UpdateHandler(api))
	r.DELETE("/users/:id", middleware.Protect(api, roles.Admin), user.DeleteHandler(api))

	// team management
	tm := r.Group("/teams", middleware.Protect(api, roles.TeamsManager))
	{
		tm.GET("/", team.MGetHandler(api))
		tm.POST("/", team.CreateHandler(api))
		tm.PATCH("/:ident", team.UpdateHandler(api))
		tm.GET("/:ident", team.GetHandler(api))
		tm.DELETE("/:ident", team.DeleteHandler(api))
		tm.PATCH("/:ident/users", team.UsersHandler(api))
	}

}
