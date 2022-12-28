package router

import (
	"strings"
	"time"

	category "github.com/fernandotsda/nemesys/api-manager/internal/alarm-category"
	alarmexp "github.com/fernandotsda/nemesys/api-manager/internal/alarm-expression"
	profile "github.com/fernandotsda/nemesys/api-manager/internal/alarm-profile"
	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/container"
	ctxmetric "github.com/fernandotsda/nemesys/api-manager/internal/contextual-metric"
	"github.com/fernandotsda/nemesys/api-manager/internal/cost"
	whitelist "github.com/fernandotsda/nemesys/api-manager/internal/counter-whitelist"
	customquery "github.com/fernandotsda/nemesys/api-manager/internal/custom-query"
	datapolicy "github.com/fernandotsda/nemesys/api-manager/internal/data-policy"
	"github.com/fernandotsda/nemesys/api-manager/internal/metric"
	metricdata "github.com/fernandotsda/nemesys/api-manager/internal/metric-data"
	"github.com/fernandotsda/nemesys/api-manager/internal/middleware"
	"github.com/fernandotsda/nemesys/api-manager/internal/refkey"
	"github.com/fernandotsda/nemesys/api-manager/internal/roles"
	"github.com/fernandotsda/nemesys/api-manager/internal/status"
	"github.com/fernandotsda/nemesys/api-manager/internal/team"
	"github.com/fernandotsda/nemesys/api-manager/internal/trap"
	"github.com/fernandotsda/nemesys/api-manager/internal/uauth"
	"github.com/fernandotsda/nemesys/api-manager/internal/user"
	"github.com/gin-contrib/cors"

	"github.com/fernandotsda/nemesys/shared/env"
	"github.com/fernandotsda/nemesys/shared/service"
	"github.com/fernandotsda/nemesys/shared/types"
	"github.com/gin-gonic/gin"
)

// Set all api routes
func Set(s service.Service) {
	api := s.(*api.API)

	gin.SetMode(gin.ReleaseMode)
	api.Log.Info("Gin mode setted to 'release'")

	router := gin.New()
	api.Router = router

	router.Use(cors.New(cors.Config{
		AllowOrigins:     strings.Split(env.APIManagerAllowOrigins, ";"),
		AllowMethods:     []string{"POST", "GET", "DELETE", "PATCH"},
		AllowHeaders:     []string{"Origin", "content-type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r := router.Group(env.APIManagerRoutesPrefix)
	r.POST("/login", middleware.Limiter(api, time.Second/2), uauth.LoginHandler(api))

	viewer := r.Group("/", middleware.Protect(api, roles.Viewer), middleware.RequestsCounter(api))
	{
		viewer.GET("/session", middleware.Protect(api, roles.Viewer), user.SessionInfoHandler(api))
		viewer.POST("/logout", middleware.Protect(api, roles.Viewer), uauth.Logout(api))
	}

	adm := r.Group("/", middleware.Protect(api, roles.Admin), middleware.RequestsCounter(api))
	{
		adm.GET("/services/status", status.GetHandler(api))
		adm.GET("/refkeys/:refkey", refkey.GetHandler(api))
		adm.GET("/cost", cost.GetCostHandler(api))
		adm.GET("/price-table", cost.GetPriceTableHandler(api))
		adm.GET("/base-plan", cost.GetBasePlanHandler(api))

		adm.POST("/metrics/data", metricdata.AddHandler(api))
	}

	master := r.Group("/", middleware.Protect(api, roles.Master))
	{
		master.PATCH("/base-plan", cost.UpdateBasePlanHandler(api))
		master.PATCH("/price-table", cost.UpdatePriceTableHandler(api))
	}

	users := r.Group("/users")
	{
		users.POST("/:userId/logout", middleware.Protect(api, roles.Admin), middleware.RequestsCounter(api), uauth.ForceLogout(api))

		users.GET("/", middleware.Protect(api, roles.TeamsManager), middleware.RequestsCounter(api), user.GetUsers(api))
		users.GET("/:userId", middleware.ProtectUser(api, roles.Admin), middleware.RequestsCounter(api), user.GetHandler(api))
		users.GET("/:userId/teams", middleware.ProtectUser(api, roles.Admin), middleware.RequestsCounter(api), user.TeamsHandler(api))
		users.POST("/", middleware.Protect(api, roles.Admin), middleware.RequestsCounter(api), user.CreateHandler(api))
		users.PATCH("/:userId", middleware.Protect(api, roles.Admin), middleware.RequestsCounter(api), user.UpdateHandler(api))
		users.DELETE("/:userId", middleware.Protect(api, roles.Admin), middleware.RequestsCounter(api), user.DeleteHandler(api))

		apikeys := users.Group("/:userId/api-keys", middleware.ProtectUser(api, roles.Admin), middleware.RequestsCounter(api))
		{
			apikeys.GET("/", user.MGetAPIKeyHandler(api))
			apikeys.POST("/", user.CreateAPIKeyHandler(api))
			apikeys.DELETE("/:apikeyId", user.DeleteAPIKeyHandler(api))
		}
	}

	teamsManagement := r.Group("/teams", middleware.Protect(api, roles.TeamsManager), middleware.RequestsCounter(api))
	{
		teamsManagement.GET("/", team.MGetHandler(api))
		teamsManagement.POST("/", team.CreateHandler(api))
		teamsManagement.PATCH("/:teamId", team.UpdateHandler(api))
		teamsManagement.DELETE("/:teamId", team.DeleteHandler(api))

		teamsManagement.GET("/:teamId/members", team.MGetMembersHandler(api))
		teamsManagement.POST("/:teamId/members", team.AddMemberHandler(api))
		teamsManagement.DELETE("/:teamId/members/:userId", team.RemoveMemberHandler(api))

		ctx := teamsManagement.Group("/:teamId/ctx")
		{
			ctx.POST("/", team.CreateContextHandler(api))
			ctx.PATCH("/:ctxId", middleware.ParseContextParams(api), team.UpdateContextHandler(api))
			ctx.DELETE("/:ctxId", middleware.ParseContextParams(api), team.DeleteContextHandler(api))
		}

		ctxMetrics := ctx.Group("/:ctxId/metrics")
		{
			ctxMetrics.POST("/", middleware.ParseContextParams(api), ctxmetric.CreateHandler(api))
			ctxMetrics.PATCH("/:ctxMetricId", middleware.ParseContextualMetricParams(api), ctxmetric.UpdateHandler(api))
			ctxMetrics.DELETE("/:ctxMetricId", middleware.ParseContextualMetricParams(api), ctxmetric.DeleteHandler(api))
		}
	}

	teams := r.Group("/teams", middleware.Protect(api, roles.Viewer), middleware.RequestsCounter(api))
	{
		teams.GET("/:teamId", middleware.TeamGuard(api, roles.Viewer, roles.TeamsManager), team.GetHandler(api))

		ctx := teams.Group("/:teamId/ctx", middleware.TeamGuard(api, roles.Viewer, roles.TeamsManager))
		{
			ctx.GET("/", team.MGetContextHandler(api))
			ctx.GET("/:ctxId", middleware.ParseContextParams(api), team.MGetContextHandler(api))
		}

		ctxMetrics := ctx.Group("/:ctxId/metrics")
		{
			ctxMetrics.GET("/", middleware.ParseContextParams(api), ctxmetric.MGet(api))
			ctxMetrics.GET("/:ctxMetricId", middleware.ParseContextualMetricParams(api), ctxmetric.Get(api))
			ctxMetrics.GET("/:ctxMetricId/alarm-history", middleware.ParseContextualMetricParams(api), middleware.MetricRequest(api), ctxmetric.AlarmHistoryHandler(api))
			ctxMetrics.GET("/:ctxMetricId/alarm-state", middleware.ParseContextualMetricParams(api), ctxmetric.GetAlarmStateHandler(api))
			ctxMetrics.POST("/:ctxMetricId/alarm-state/recognize", middleware.ParseContextualMetricParams(api), ctxmetric.RecognizeAlarmStateHandler(api))
			ctxMetrics.POST("/:ctxMetricId/alarm-state/resolve", middleware.ParseContextualMetricParams(api), ctxmetric.ResolveAlarmStateHandler(api))
		}
	}

	dp := r.Group("/data-policies", middleware.Protect(api, roles.Admin), middleware.RequestsCounter(api))
	{
		dp.GET("/", datapolicy.MGetHandler(api))
		dp.GET("/:dpId", datapolicy.GetHandler(api))
		dp.POST("/", datapolicy.CreateHandler(api))
		dp.PATCH("/:dpId", datapolicy.UpdateHandler(api))
		dp.DELETE("/:dpId", datapolicy.DeleteHandler(api))
	}

	basic := r.Group("/containers/basics", middleware.Protect(api, roles.Admin), middleware.RequestsCounter(api))
	{
		basic.GET("/", container.GetBasicContainersHandlers(api))
		basic.GET("/:containerId", container.GetBasicHandler(api))
		basic.POST("/", container.CreateBasicHandler(api))
		basic.PATCH("/:containerId", container.UpdateBasicHandler(api))
		basic.DELETE("/:containerId", container.DeleteHandler(api))

		metrics := basic.Group("/:containerId/metrics")
		{
			setGetAlarmExpressions(api, metrics)

			metrics.GET("/", metric.MGetBasicHandler(api))
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

	SNMPv2c := r.Group("/containers/snmpv2c", middleware.Protect(api, roles.Admin), middleware.RequestsCounter(api))
	{
		SNMPv2c.GET("/", container.GetSNMPv2cContainers(api))
		SNMPv2c.GET("/:containerId", container.GetSNMPv2cHandler(api))
		SNMPv2c.POST("/", container.CreateSNMPv2cHandler(api))
		SNMPv2c.PATCH("/:containerId", container.UpdateSNMPv2cHandler(api))
		SNMPv2c.DELETE("/:containerId", container.DeleteHandler(api))

		metrics := SNMPv2c.Group("/:containerId/metrics")
		{
			setGetAlarmExpressions(api, metrics)

			metrics.GET("/", metric.MGetSNMPv2cHandler(api))
			metrics.GET("/:metricId", metric.GetSNMPv2cHandler(api))
			metrics.POST("/", metric.CreateSNMPv2cHandler(api))
			metrics.PATCH("/:metricId", metric.UpdateSNMPv2cHandler(api))
			metrics.DELETE("/:metricId", metric.DeleteHandler(api))
		}
	}

	flexLegacy := r.Group("/containers/flex-legacy", middleware.Protect(api, roles.Admin), middleware.RequestsCounter(api))
	{
		flexLegacy.GET("/", container.GetFlexLegacyContainersHandler(api))
		flexLegacy.GET("/:containerId", container.GetFlexLegacyHandler(api))
		flexLegacy.POST("/", container.CreateFlexLegacy(api))
		flexLegacy.PATCH("/:containerId", container.UpdateFlexLegacy(api))
		flexLegacy.DELETE(":containerId", container.DeleteHandler(api))

		metrics := flexLegacy.Group("/:containerId/metrics")
		{
			setGetAlarmExpressions(api, metrics)

			metrics.GET("/", metric.MGetFlexLegacyHandler(api))
			metrics.GET("/:metricId", metric.GetFlexLegacyHandler(api))
			metrics.POST("/", metric.CreateFlexLegacyHandler(api))
			metrics.PATCH("/:metricId", metric.UpdateFlexLegacyHandler(api))
			metrics.DELETE("/:metricId", metric.DeleteHandler(api))
		}
	}

	customQuery := r.Group("/custom-queries")
	{
		customQuery.GET("/", middleware.Protect(api, roles.Viewer), middleware.RequestsCounter(api), customquery.MGetHandler(api))
		customQuery.GET("/:cqId", middleware.Protect(api, roles.Viewer), middleware.RequestsCounter(api), customquery.GetHandler(api))
		customQuery.POST("/", middleware.Protect(api, roles.TeamsManager), middleware.RequestsCounter(api), customquery.CreateHandler(api))
		customQuery.PATCH("/:cqId", middleware.Protect(api, roles.TeamsManager), middleware.RequestsCounter(api), customquery.UpdateHandler(api))
		customQuery.DELETE("/:cqId", middleware.Protect(api, roles.TeamsManager), middleware.RequestsCounter(api), customquery.DeleteHandler(api))
	}

	alarmProfile := r.Group("/alarm/profiles", middleware.Protect(api, roles.Admin), middleware.RequestsCounter(api))
	{
		alarmProfile.GET("/", profile.MGetHandler(api))
		alarmProfile.GET("/:profileId", profile.GetHandler(api))
		alarmProfile.POST("/", profile.CreateHandler(api))
		alarmProfile.PATCH("/:profileId", profile.UpdateHandler(api))
		alarmProfile.DELETE("/:profileId", profile.DeleteHandler(api))

		emails := alarmProfile.Group("/:profileId/emails")
		{
			emails.GET("/", profile.GetEmailsHandler(api))
			emails.POST("/", profile.CreateEmailHandler(api))
			emails.DELETE("/:emailId", profile.DeleteEmailHandler(api))
		}

		category := alarmProfile.Group("/:profileId/categories")
		{
			category.GET("/", profile.GetCategoriesHandler(api))
			category.POST("/", profile.AddCategoryHandler(api))
			category.DELETE("/:categoryId", profile.RemoveCategoryHandler(api))
		}
	}

	alarmCategory := r.Group("/alarm/categories", middleware.Protect(api, roles.Admin), middleware.RequestsCounter(api))
	{
		alarmCategory.GET("/", category.MGetHandler(api))
		alarmCategory.GET("/:categoryId", category.GetHandler(api))
		alarmCategory.POST("/", category.CreateHandler(api))
		alarmCategory.PATCH("/:categoryId", category.UpdateHandler(api))
		alarmCategory.DELETE("/:categoryId", category.DeleteHandler(api))
	}

	alarmExpression := r.Group("/alarm/expressions", middleware.Protect(api, roles.Admin), middleware.RequestsCounter(api))
	{
		alarmExpression.GET("/", alarmexp.MGetHandler(api))
		alarmExpression.POST("/", alarmexp.CreateHandler(api))
		alarmExpression.PATCH("/:expressionId", alarmexp.UpdateHandler(api))
		alarmExpression.DELETE("/:expressionId", alarmexp.DeleteHandler(api))
		alarmExpression.POST("/:expressionId/metrics", alarmexp.CreateMetricRelationHandler(api))
		alarmExpression.DELETE("/:expressionId/metrics/:metricId", alarmexp.DeleteMetricRelationHandler(api))
	}

	trapRelations := r.Group("/alarm/trap-relations/", middleware.Protect(api, roles.Admin), middleware.RequestsCounter(api))
	{
		trapRelations.GET("/", category.GetTrapRelationsHandler(api))
		trapRelations.POST("/", category.CreateTrapRelationHandler(api))
		trapRelations.DELETE("/:trapId", category.DeleteTrapRelationHandler(api))
	}

	requestWhitelist := r.Group("request-count/whitelist/members", middleware.Protect(api, roles.Master))
	{
		requestWhitelist.GET("/", whitelist.GetHandler(api))
		requestWhitelist.POST("/", whitelist.CreateHandler(api))
		requestWhitelist.DELETE("/:userId", whitelist.DeleteHandler(api))
	}

	trapListeners := r.Group("/trap-listeners", middleware.Protect(api, roles.Admin))
	{
		trapListeners.GET("/", trap.MGetHandler(api))
		trapListeners.POST("/", trap.CreateHandler(api))
		trapListeners.PATCH("/:listenerId", trap.UpdateHandler(api))
		trapListeners.DELETE("/:listenerId", trap.DeleteHandler(api))
	}

	// metric data
	{
		r.GET("/teams/:teamId/ctx/:ctxId/metrics/:ctxMetricId/data",
			middleware.Protect(api, roles.Viewer),
			middleware.RealtimeDataRequestsCounter(api),
			middleware.ParseContextualMetricParams(api),
			middleware.MetricRequest(api),
			ctxmetric.DataHandler(api),
		)
		r.GET("/teams/:teamId/ctx/:ctxId/metrics/:ctxMetricId/data/history",
			middleware.Protect(api, roles.Viewer),
			middleware.DataHistoryRequestsCounter(api),
			middleware.ParseContextualMetricParams(api),
			middleware.MetricRequest(api),
			ctxmetric.QueryDataHandler(api),
		)
	}
}

func setGetAlarmExpressions(api *api.API, r *gin.RouterGroup) {
	r.GET("/:metricId/alarm-expressions", metric.GetAlarmExpressions(api))
}
