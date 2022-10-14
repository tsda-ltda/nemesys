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

	// teams
	teams := r.Group("/teams", middleware.Protect(api, roles.Viewer))
	{
		teams.GET("/", team.UserTeamsHandler(api))
	}

	// users config
	usersConfig := r.Group("/config/users")
	{
		usersConfig.GET("/", middleware.Protect(api, roles.TeamsManager), user.MGetHandler(api))
		usersConfig.GET("/:id", middleware.ProtectUser(api, roles.Admin), user.GetHandler(api))
		usersConfig.POST("/", middleware.Protect(api, roles.Admin), user.CreateHandler(api))
		usersConfig.PATCH("/:id", middleware.Protect(api, roles.Admin), user.UpdateHandler(api))
		usersConfig.DELETE("/:id", middleware.Protect(api, roles.Admin), user.DeleteHandler(api))
	}

	// teams config
	teamConfig := r.Group("/config/teams", middleware.Protect(api, roles.TeamsManager))
	{
		teamConfig.GET("/", team.MGetHandler(api))
		teamConfig.GET("/:id", team.GetHandler(api))
		teamConfig.POST("/", team.CreateHandler(api))
		teamConfig.POST("/:id/users", team.AddUserHandler(api))
		teamConfig.POST("/:id/contexts", team.CreateContextHandler(api))
		teamConfig.PATCH("/:id", team.UpdateHandler(api))
		teamConfig.DELETE("/:id", team.DeleteHandler(api))
		teamConfig.DELETE("/:id/users/:userId", team.RemoveUserHandler(api))
		teamConfig.DELETE("/:id/contexts/:contextId", team.DeleteContextHandler(api))
	}

	// data-policies
	dp := r.Group("/config/data-policies", middleware.Protect(api, roles.Master))
	{
		dp.GET("/", datapolicy.MGetHandler(api))
		dp.POST("/", datapolicy.CreateHandler(api))
		dp.PATCH("/:id", datapolicy.UpdateHandler(api))
		dp.DELETE("/:id", datapolicy.DeleteHandler(api))
	}
}
