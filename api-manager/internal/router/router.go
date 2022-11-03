package router

import (
	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/container"
	ctxmetric "github.com/fernandotsda/nemesys/api-manager/internal/contextual-metric"
	datapolicy "github.com/fernandotsda/nemesys/api-manager/internal/data-policy"
	"github.com/fernandotsda/nemesys/api-manager/internal/metric"
	"github.com/fernandotsda/nemesys/api-manager/internal/middleware"
	"github.com/fernandotsda/nemesys/api-manager/internal/roles"
	"github.com/fernandotsda/nemesys/api-manager/internal/team"
	"github.com/fernandotsda/nemesys/api-manager/internal/uauth"
	"github.com/fernandotsda/nemesys/api-manager/internal/user"

	"github.com/fernandotsda/nemesys/shared/env"
	"github.com/fernandotsda/nemesys/shared/types"
)

// Set all api routes
func Set(api *api.API) {
	// get api routes
	r := api.Router.Group(env.APIManagerRoutesPrefix)

	// user authentication
	r.POST("/login", uauth.LoginHandler(api))
	r.POST("/logout", middleware.Protect(api, roles.Viewer), uauth.Logout(api))

	// users config
	users := r.Group("/users")
	{
		// force logout
		users.POST("/:id/logout", middleware.Protect(api, roles.Admin), uauth.ForceLogout(api))

		// user
		users.GET("/", middleware.Protect(api, roles.TeamsManager), user.MGetHandler(api))
		users.GET("/:id", middleware.ProtectUser(api, roles.Admin), user.GetHandler(api))
		users.POST("/", middleware.Protect(api, roles.Admin), user.CreateHandler(api))
		users.PATCH("/:id", middleware.Protect(api, roles.Admin), user.UpdateHandler(api))
		users.DELETE("/:id", middleware.Protect(api, roles.Admin), user.DeleteHandler(api))

		// user's teams
		users.GET("/:id/teams", middleware.ProtectUser(api, roles.Admin), user.TeamsHandler(api))
	}

	// teams and context config
	teams := r.Group("config/teams", middleware.Protect(api, roles.TeamsManager))
	{
		// teams
		teams.GET("/", team.MGetHandler(api))
		teams.GET("/:id", team.GetHandler(api))
		teams.POST("/", team.CreateHandler(api))
		teams.PATCH("/:id", team.UpdateHandler(api))
		teams.DELETE("/:id", team.DeleteHandler(api))

		// members
		teams.GET("/:id/members", team.MGetMembersHandler(api))
		teams.POST("/:id/members", team.AddMemberHandler(api))
		teams.DELETE("/:id/members/:userId", team.RemoveMemberHandler(api))

		// contexts
		teams.GET("/:id/ctx", team.MGetContextHandler(api))
		teams.GET("/:id/ctx/:ctxId", team.MGetContextHandler(api))
		teams.POST("/:id/ctx", team.CreateContextHandler(api))
		teams.PATCH("/:id/ctx/:ctxId", team.UpdateContextHandler(api))
		teams.DELETE("/:id/ctx/:ctxId", team.DeleteContextHandler(api))

		// contextual metrics
		teams.GET("/:id/ctx/:ctxId/metrics", ctxmetric.MGet(api))
		teams.GET("/:id/ctx/:ctxId/metrics/:metricId", ctxmetric.Get(api))
		teams.POST("/:id/ctx/:ctxId/metrics", ctxmetric.CreateHandler(api))
		teams.PATCH("/:id/ctx/:ctxId/metrics/:metricId", ctxmetric.UpdateHandler(api))
		teams.DELETE("/:id/ctx/:ctxId/metrics/:metricId", ctxmetric.DeleteHandler(api))
	}

	// data-policies
	dp := r.Group("/data-policies", middleware.Protect(api, roles.Master))
	{
		dp.GET("/", datapolicy.MGetHandler(api))
		dp.POST("/", datapolicy.CreateHandler(api))
		dp.PATCH("/:id", datapolicy.UpdateHandler(api))
		dp.DELETE("/:id", datapolicy.DeleteHandler(api))
	}

	// snmp container and metrics
	snmp := r.Group("/containers/snmp", middleware.Protect(api, roles.Admin))
	{
		// container
		snmp.GET("/", container.MGet(api, types.CTSNMP))
		snmp.GET("/:containerId", container.GetSNMPHandler(api))
		snmp.POST("/", container.CreateSNMPHandler(api))
		snmp.PATCH("/:containerId", container.UpdateSNMPHandler(api))
		snmp.DELETE("/:containerId", container.DeleteHandler(api))

		// metric
		snmp.GET("/:containerId/metrics", metric.MGet(api, types.CTSNMP))
		snmp.GET("/:containerId/metrics/:metricId", metric.GetSNMPHandler(api))
		snmp.POST("/:containerId/metrics", metric.CreateSNMPHandler(api))
		snmp.PATCH("/:containerId/metrics/:metricId", metric.UpdateSNMPHandler(api))
		snmp.DELETE("/:containerId/metrics/:metricId", metric.DeleteHandler(api))
	}

	data := r.Group("/teams/:teamIdent/ctx/:ctxIdent/metrics")
	{
		data.GET("/:metricIdent/data", ctxmetric.DataHandler(api))
	}
}
