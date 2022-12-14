package metric

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/gin-gonic/gin"
)

// Get a basic metric.
// Responses:
//   - 400 If invalid params.
//   - 404 If not found.
//   - 200 If succeeded.
func GetBasicHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		id, err := strconv.ParseInt(c.Param("metricId"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}

		exists, metric, err := api.PG.GetBasicMetric(ctx, id)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to get flex legacy metric", logger.ErrField(err))
			return
		}
		if !exists {
			c.JSON(http.StatusNotFound, tools.MsgRes(tools.MsgMetricNotFound))
			return
		}

		c.JSON(http.StatusOK, tools.DataRes(metric))
	}
}
