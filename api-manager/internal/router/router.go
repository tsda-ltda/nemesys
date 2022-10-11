package router

import (
	datapolicy "github.com/fernandotsda/nemesys/api-manager/internal/data-policy"
	"github.com/fernandotsda/nemesys/api-manager/internal/middleware"
	"github.com/fernandotsda/nemesys/api-manager/internal/roles"
	"github.com/fernandotsda/nemesys/api-manager/internal/team"
	"github.com/fernandotsda/nemesys/api-manager/internal/user"
	"github.com/fernandotsda/nemesys/shared/env"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/uauth"
)

// Set all api routes
func Set(api *api.API) {
	// get api routes
	r := api.Router.Group(env.APIManagerRoutesPrefix)

	// User authentication
	r.POST("/login", uauth.LoginHandler(api))
	r.POST("/logout", middleware.Protect(api, roles.Viewer), uauth.Logout(api))
	r.POST("/users/:id/logout", middleware.Protect(api, roles.Admin), uauth.ForceLogout(api))

	// users
	users := r.Group("/users")
	{
		users.GET("/", middleware.Protect(api, roles.TeamsManager), user.MGetHandler(api))
		users.POST("/", middleware.Protect(api, roles.Admin), user.CreateHandler(api))
		users.GET("/:id", middleware.ProtectUser(api, roles.Admin), user.GetHandler(api))
		users.PATCH("/:id", middleware.Protect(api, roles.Admin), user.UpdateHandler(api))
		users.DELETE("/:id", middleware.Protect(api, roles.Admin), user.DeleteHandler(api))
	}

	// teams
	teams := r.Group("/teams", middleware.Protect(api, roles.Viewer))
	{
		teams.GET("/", team.UserTeamsHandler(api))
	}

	// teams config
	teamConfig := r.Group("/config/teams", middleware.Protect(api, roles.TeamsManager))
	{
		teamConfig.GET("/", team.MGetHandler(api))
		teamConfig.POST("/", team.CreateHandler(api))
		teamConfig.PATCH("/:ident", team.UpdateHandler(api))
		teamConfig.GET("/:ident", team.GetHandler(api))
		teamConfig.DELETE("/:ident", team.DeleteHandler(api))
		teamConfig.POST("/:ident/users", team.AddUserHandler(api))
		teamConfig.DELETE("/:ident/users/:userId", team.RemoveUserHandler(api))
	}

	// data-policies
	dp := r.Group("/config/data-policies", middleware.Protect(api, roles.Master))
	{
		dp.GET("/", datapolicy.MGetHandler(api))
		dp.POST("/", datapolicy.CreateHandler(api))
		dp.DELETE("/:id", datapolicy.DeleteHandler(api))
		dp.PATCH("/:id", datapolicy.UpdateHandler(api))
	}

}
