package metric

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/types"
	"github.com/gin-gonic/gin"
)

// Get a Flex Legacy metric.
// Responses:
//   - 400 If invalid params.
//   - 404 If not found.
//   - 200 If succeeded.
func GetFlexLegacyHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// metric id
		metricId, err := strconv.ParseInt(c.Param("metricId"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		// get metric base information
		base, err := api.PgConn.Metrics.Get(ctx, metricId)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to get metric", logger.ErrField(err))
			return
		}
		base.Metric.ContainerType = types.CTFlexLegacy

		// check if exists
		if !base.Exists {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgMetricNotFound))
			return
		}

		// get metric
		protocol, err := api.PgConn.FlexLegacyMetrics.Get(ctx, metricId)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to get flex legacy metric", logger.ErrField(err))
			return
		}

		// check if exists
		if !protocol.Exists {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgMetricNotFound))
			return
		}

		metric := models.Metric[models.FlexLegacyMetric]{
			Base:     base.Metric,
			Protocol: protocol.Metric,
		}

		c.JSON(http.StatusOK, metric)
	}
}
