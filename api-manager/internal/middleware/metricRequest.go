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

		// get contextual metric ctxMetricId
		ctxMetricId, err := strconv.ParseInt(c.Param("metricId"), 10, 64)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		// get metric request
		r, err := api.PgConn.ContextualMetrics.GetMetricEnabledAndRequestById(ctx, ctxMetricId)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			api.Log.Error("fail to get contextual metric, team and context id on database", logger.ErrField(err))
			return
		}

		// check if exists
		if !r.Exists {
			c.AbortWithStatusJSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgContextualMetricNotFound))
			return
		}

		// check if is enabled
		if !r.Enabled {
			c.AbortWithStatusJSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgMetricDisabled))
			return
		}

		// save metric request
		c.Set("metric_request", r.MetricRequest)
		c.Next()
	}
}
