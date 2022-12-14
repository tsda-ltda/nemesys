package middleware

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/gin-gonic/gin"
)

// MetricRequest fetchs the metric request information on database and sets in
// the request contexts. Also checks if the metric is enabled.
func MetricRequest(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		ctxMetricId, err := strconv.ParseInt(c.Param("ctxMetricId"), 10, 64)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}

		cacheR, err := api.Cache.GetMetricRequest(ctx, ctxMetricId)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.AbortWithStatus(http.StatusInternalServerError)
			api.Log.Error("Fail to get metric request on cache", logger.ErrField(err))
			return
		}
		if cacheR.Exists {
			c.Set("metric_request", cacheR.Request)
			c.Next()
			return
		}

		r, err := api.PG.GetMetricRequestByContextualMetric(ctx, ctxMetricId)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.AbortWithStatus(http.StatusInternalServerError)
			api.Log.Error("Fail to get contextual metric, team and context id on database", logger.ErrField(err))
			return
		}
		if !r.Exists {
			c.AbortWithStatusJSON(http.StatusBadRequest, tools.MsgRes(tools.MsgContextualMetricNotFound))
			return
		}
		if !r.Enabled {
			c.AbortWithStatusJSON(http.StatusBadRequest, tools.MsgRes(tools.MsgMetricDisabled))
			return
		}

		err = api.Cache.SetMetricRequest(ctx, ctxMetricId, r.MetricRequest)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.AbortWithStatus(http.StatusInternalServerError)
			api.Log.Error("Fail to set metric request on cache", logger.ErrField(err))
			return
		}

		c.Set("metric_request", r.MetricRequest)
		c.Next()
	}
}
