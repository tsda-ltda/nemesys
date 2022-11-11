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
	teams := r.Group("/teams", middleware.Protect(api, roles.TeamsManager))
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
		ctx := teams.Group("/:id/ctx")
		{
			ctx.GET("/", team.MGetContextHandler(api))
			ctx.GET("/:ctxId", middleware.ParseContextParams(api), team.MGetContextHandler(api))
			ctx.POST("/", team.CreateContextHandler(api))
			ctx.PATCH("/:ctxId", middleware.ParseContextParams(api), team.UpdateContextHandler(api))
			ctx.DELETE("/:ctxId", middleware.ParseContextParams(api), team.DeleteContextHandler(api))
		}

		// contextual metrics
		ctxMetric := ctx.Group("/:ctxId/metrics")
		{
			ctxMetric.GET("/", middleware.ParseContextParams(api), ctxmetric.MGet(api))
			ctxMetric.GET("/:metricId", middleware.ParseContextualMetricParams(api), ctxmetric.Get(api))
			ctxMetric.GET("/:metricId/data", middleware.ParseContextualMetricParams(api), middleware.MetricRequest(api), ctxmetric.DataHandler(api))
			ctxMetric.POST("/", middleware.ParseContextParams(api), ctxmetric.CreateHandler(api))
			ctxMetric.PATCH("/:metricId", middleware.ParseContextualMetricParams(api), ctxmetric.UpdateHandler(api))
			ctxMetric.DELETE("/:metricId", middleware.ParseContextualMetricParams(api), ctxmetric.DeleteHandler(api))
		}
	}

	// data-policies
	dp := r.Group("/data-policies", middleware.Protect(api, roles.Master))
	{
		dp.GET("/", datapolicy.MGetHandler(api))
		dp.GET("/:id", datapolicy.GetHandler(api))
		dp.POST("/", datapolicy.CreateHandler(api))
		dp.PATCH("/:id", datapolicy.UpdateHandler(api))
		dp.DELETE("/:id", datapolicy.DeleteHandler(api))
	}

	// SNMPv2c container and metrics
	SNMPv2c := r.Group("/containers/snmpv2c", middleware.Protect(api, roles.Admin))
	{
		// container
		SNMPv2c.GET("/", container.MGet(api, types.CTSNMPv2c))
		SNMPv2c.GET("/:containerId", container.GetSNMPv2cHandler(api))
		SNMPv2c.POST("/", container.CreateSNMPv2cHandler(api))
		SNMPv2c.PATCH("/:containerId", container.UpdateSNMPv2cHandler(api))
		SNMPv2c.DELETE("/:containerId", container.DeleteHandler(api))

		// metric
		SNMPv2c.GET("/:containerId/metrics", metric.MGet(api, types.CTSNMPv2c))
		SNMPv2c.GET("/:containerId/metrics/:metricId", metric.GetSNMPv2cHandler(api))
		SNMPv2c.POST("/:containerId/metrics", metric.CreateSNMPv2cHandler(api))
		SNMPv2c.PATCH("/:containerId/metrics/:metricId", metric.UpdateSNMPv2cHandler(api))
		SNMPv2c.DELETE("/:containerId/metrics/:metricId", metric.DeleteHandler(api))
	}
}
