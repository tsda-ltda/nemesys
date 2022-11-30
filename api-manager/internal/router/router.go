package router

import (
	"time"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/container"
	ctxmetric "github.com/fernandotsda/nemesys/api-manager/internal/contextual-metric"
	customquery "github.com/fernandotsda/nemesys/api-manager/internal/custom-query"
	datapolicy "github.com/fernandotsda/nemesys/api-manager/internal/data-policy"
	"github.com/fernandotsda/nemesys/api-manager/internal/metric"
	metricdata "github.com/fernandotsda/nemesys/api-manager/internal/metric-data"
	"github.com/fernandotsda/nemesys/api-manager/internal/middleware"
	"github.com/fernandotsda/nemesys/api-manager/internal/refkey"
	"github.com/fernandotsda/nemesys/api-manager/internal/roles"
	"github.com/fernandotsda/nemesys/api-manager/internal/team"
	"github.com/fernandotsda/nemesys/api-manager/internal/uauth"
	"github.com/fernandotsda/nemesys/api-manager/internal/user"

	"github.com/fernandotsda/nemesys/shared/env"
	"github.com/fernandotsda/nemesys/shared/service"
	"github.com/fernandotsda/nemesys/shared/types"
)

// Set all api routes
func Set(s service.Service) {
	api := s.(*api.API)

	r := api.Router.Group(env.APIManagerRoutesPrefix)

	r.POST("/login", middleware.Limiter(api, time.Second/2), uauth.LoginHandler(api))
	r.POST("/logout", middleware.Protect(api, roles.Viewer), uauth.Logout(api))

	global := r.Group("/global", middleware.Protect(api, roles.Admin))
	{
		global.GET("/refkeys/:refkey", refkey.GetHandler(api))
		global.POST("/metrics/data", metricdata.AddHandler(api))
	}

	users := r.Group("/users")
	{
		users.POST("/:userId/logout", middleware.Protect(api, roles.Admin), uauth.ForceLogout(api))

		users.GET("/", middleware.Protect(api, roles.TeamsManager), user.MGetHandler(api))
		users.GET("/:userId", middleware.ProtectUser(api, roles.Admin), user.GetHandler(api))
		users.GET("/:userId/teams", middleware.ProtectUser(api, roles.Admin), user.TeamsHandler(api))
		users.POST("/", middleware.Protect(api, roles.Admin), user.CreateHandler(api))
		users.PATCH("/:userId", middleware.Protect(api, roles.Admin), user.UpdateHandler(api))
		users.DELETE("/:userId", middleware.Protect(api, roles.Admin), user.DeleteHandler(api))

		apikeys := users.Group("/:userId/api-keys", middleware.ProtectUser(api, roles.Admin))
		{
			apikeys.GET("/", user.MGetAPIKeyHandler(api))
			apikeys.POST("/", user.CreateAPIKeyHandler(api))
			apikeys.DELETE("/:apikeyId", user.DeleteAPIKeyHandler(api))
		}
	}

	teams := r.Group("/teams", middleware.Protect(api, roles.TeamsManager))
	{
		teams.GET("/", team.MGetHandler(api))
		teams.GET("/:teamId", team.GetHandler(api))
		teams.POST("/", team.CreateHandler(api))
		teams.PATCH("/:teamId", team.UpdateHandler(api))
		teams.DELETE("/:teamId", team.DeleteHandler(api))

		teams.GET("/:teamId/members", team.MGetMembersHandler(api))
		teams.POST("/:teamId/members", team.AddMemberHandler(api))
		teams.DELETE("/:teamId/members/:userId", team.RemoveMemberHandler(api))

		ctx := teams.Group("/:teamId/ctx")
		{
			ctx.GET("/", team.MGetContextHandler(api))
			ctx.GET("/:ctxId", middleware.ParseContextParams(api), team.MGetContextHandler(api))
			ctx.POST("/", team.CreateContextHandler(api))
			ctx.PATCH("/:ctxId", middleware.ParseContextParams(api), team.UpdateContextHandler(api))
			ctx.DELETE("/:ctxId", middleware.ParseContextParams(api), team.DeleteContextHandler(api))
		}

		ctxMetrics := ctx.Group("/:ctxId/metrics")
		{
			ctxMetrics.GET("/", middleware.ParseContextParams(api), ctxmetric.MGet(api))
			ctxMetrics.GET("/:metricId", middleware.ParseContextualMetricParams(api), ctxmetric.Get(api))
			ctxMetrics.POST("/", middleware.ParseContextParams(api), ctxmetric.CreateHandler(api))
			ctxMetrics.PATCH("/:metricId", middleware.ParseContextualMetricParams(api), ctxmetric.UpdateHandler(api))
			ctxMetrics.DELETE("/:metricId", middleware.ParseContextualMetricParams(api), ctxmetric.DeleteHandler(api))
		}
	}

	dp := r.Group("/data-policies", middleware.Protect(api, roles.Master))
	{
		dp.GET("/", datapolicy.MGetHandler(api))
		dp.GET("/:dpId", datapolicy.GetHandler(api))
		dp.POST("/", datapolicy.CreateHandler(api))
		dp.PATCH("/:dpId", datapolicy.UpdateHandler(api))
		dp.DELETE("/:dpId", datapolicy.DeleteHandler(api))
	}

	basic := r.Group("/containers/basics", middleware.Protect(api, roles.Admin))
	{
		basic.GET("/", container.MGet(api, types.CTBasic))
		basic.GET("/:containerId", container.GetBasicHandler(api))
		basic.POST("/", container.CreateBasicHandler(api))
		basic.PATCH("/:containerId", container.UpdateBasicHandler(api))
		basic.DELETE("/:containerId", container.DeleteHandler(api))

		metrics := basic.Group("/:containerId/metrics")
		{
			metrics.GET("/", metric.MGet(api, types.CTBasic))
			metrics.GET("/:metricId", metric.GetBasicHandler(api))
			metrics.POST("/", metric.CreateBasicHandler(api))
			metrics.PATCH("/:metricId", metric.UpdateBasicHandler(api))
			metrics.DELETE("/:metricId", metric.DeleteHandler(api))

			refkeys := metrics.Group("/:metricId/refkeys")
			{
				refkeys.GET("/", refkey.MGetHandler(api))
				refkeys.POST("/", refkey.CreateHandler(api, types.CTBasic))
				refkeys.PATCH("/:refkeyId", refkey.UpdateHandler(api, types.CTBasic))
				refkeys.DELETE("/:refkeyId", refkey.DeleteHandler(api))
			}
		}
	}

	SNMPv2c := r.Group("/containers/snmpv2c", middleware.Protect(api, roles.Admin))
	{
		SNMPv2c.GET("/", container.MGet(api, types.CTSNMPv2c))
		SNMPv2c.GET("/:containerId", container.GetSNMPv2cHandler(api))
		SNMPv2c.POST("/", container.CreateSNMPv2cHandler(api))
		SNMPv2c.PATCH("/:containerId", container.UpdateSNMPv2cHandler(api))
		SNMPv2c.DELETE("/:containerId", container.DeleteHandler(api))

		metrics := SNMPv2c.Group("/:containerId/metrics")
		{
			metrics.GET("/", metric.MGet(api, types.CTSNMPv2c))
			metrics.GET("/:metricId", metric.GetSNMPv2cHandler(api))
			metrics.POST("/", metric.CreateSNMPv2cHandler(api))
			metrics.PATCH("/:metricId", metric.UpdateSNMPv2cHandler(api))
			metrics.DELETE("/:metricId", metric.DeleteHandler(api))
		}
	}

	flexLegacy := r.Group("/containers/flex-legacy", middleware.Protect(api, roles.Admin))
	{
		flexLegacy.GET("/", container.MGet(api, types.CTFlexLegacy))
		flexLegacy.GET("/:containerId", container.GetFlexLegacyHandler(api))
		flexLegacy.POST("/", container.CreateFlexLegacy(api))
		flexLegacy.PATCH("/:containerId", container.UpdateFlexLegacy(api))
		flexLegacy.DELETE(":containerId", container.DeleteHandler(api))

		metrics := flexLegacy.Group("/:containerId/metrics")
		{
			metrics.GET("/", metric.MGet(api, types.CTFlexLegacy))
			metrics.GET("/:metricId", metric.GetFlexLegacyHandler(api))
			metrics.POST("/", metric.CreateFlexLegacyHandler(api))
			metrics.PATCH("/:metricId", metric.UpdateFlexLegacyHandler(api))
			metrics.DELETE("/:metricId", metric.DeleteHandler(api))
		}
	}

	customQuery := r.Group("/custom-queries")
	{
		customQuery.GET("/", middleware.Protect(api, roles.Viewer), customquery.MGetHandler(api))
		customQuery.GET("/:cqId", middleware.Protect(api, roles.Viewer), customquery.GetHandler(api))
		customQuery.POST("/", middleware.Protect(api, roles.TeamsManager), customquery.CreateHandler(api))
		customQuery.PATCH("/:cqId", middleware.Protect(api, roles.TeamsManager), customquery.UpdateHandler(api))
		customQuery.DELETE("/:cqId", middleware.Protect(api, roles.TeamsManager), customquery.DeleteHandler(api))
	}

	{
		r.GET("/teams/:teamId/ctx/:ctxId/metrics/:metricId/data",
			middleware.Limiter(api, time.Millisecond*300),
			middleware.Protect(api, roles.Viewer),
			middleware.ParseContextualMetricParams(api),
			middleware.MetricRequest(api),
			ctxmetric.DataHandler(api),
		)
		r.GET("/teams/:teamId/ctx/:ctxId/metrics/:metricId/data/history",
			middleware.Limiter(api, time.Millisecond*650),
			middleware.Protect(api, roles.Viewer),
			middleware.ParseContextualMetricParams(api),
			middleware.MetricRequest(api),
			ctxmetric.QueryDataHandler(api),
		)
	}
}
