package router

import (
	"github.com/fernandotsda/nemesys/api-manager/internal/container"
	datapolicy "github.com/fernandotsda/nemesys/api-manager/internal/data-policy"
	"github.com/fernandotsda/nemesys/api-manager/internal/metric"
	"github.com/fernandotsda/nemesys/api-manager/internal/middleware"
	"github.com/fernandotsda/nemesys/api-manager/internal/roles"
	"github.com/fernandotsda/nemesys/api-manager/internal/team"
	"github.com/fernandotsda/nemesys/api-manager/internal/user"
	"github.com/fernandotsda/nemesys/shared/env"
	"github.com/fernandotsda/nemesys/shared/types"

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
		teamConfig.GET("/:id/members", team.MGetMembersHandler(api))
		teamConfig.GET("/:id/ctxs", team.MGetContextHandler(api))
		teamConfig.POST("/", team.CreateHandler(api))
		teamConfig.POST("/:id/members", team.AddMemberHandler(api))
		teamConfig.POST("/:id/ctxs", team.CreateContextHandler(api))
		teamConfig.PATCH("/:id", team.UpdateHandler(api))
		teamConfig.DELETE("/:id", team.DeleteHandler(api))
		teamConfig.DELETE("/:id/members/:userId", team.RemoveMemberHandler(api))
		teamConfig.DELETE("/:id/ctxs/:ctxId", team.DeleteContextHandler(api))
	}

	// data-policies
	dp := r.Group("/config/data-policies", middleware.Protect(api, roles.Master))
	{
		dp.GET("/", datapolicy.MGetHandler(api))
		dp.POST("/", datapolicy.CreateHandler(api))
		dp.PATCH("/:id", datapolicy.UpdateHandler(api))
		dp.DELETE("/:id", datapolicy.DeleteHandler(api))
	}

	// snmpContainer container
	snmpContainer := r.Group("/config/containers/snmp", middleware.Protect(api, roles.Admin))
	{
		snmpContainer.GET("/", container.MGet(api, types.CTSNMP))
		snmpContainer.GET("/:id", container.GetSNMPHandler(api))
		snmpContainer.POST("/", container.CreateSNMPHandler(api))
		snmpContainer.PATCH("/:id", container.UpdateSNMPHandler(api))
		snmpContainer.DELETE("/:id", container.DeleteHandler(api))
	}

	// snmp metrics
	snmpMetric := r.Group("/config/metrics/snmp", middleware.Protect(api, roles.Admin))
	{
		snmpMetric.GET("/", metric.MGet(api, types.CTSNMP))
		snmpMetric.GET("/:id", metric.GetSNMPHandler(api))
		snmpMetric.POST("/", metric.CreateSNMPHandler(api))
		snmpMetric.PATCH("/:id", metric.UpdateSNMPHandler(api))
		snmpMetric.DELETE("/:id", metric.DeleteHandler(api))
	}
}
