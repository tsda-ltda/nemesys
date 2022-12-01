package alarmexp

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/gin-gonic/gin"
)

// Deletes a alarm expression.
// Responses:
//   - 400 If invalid params.
//   - 404 If not found.
//   - 204 If succeeded.
func DeleteHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		rawMetricId := c.Param("metricId")
		metricId, err := strconv.ParseInt(rawMetricId, 0, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}
		exists, err := api.PG.DeleteAlarmExpression(ctx, metricId)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to delete alarm expression", logger.ErrField(err))
			return
		}
		if !exists {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgAlarmExpressionNotFound))
			return
		}
		api.Log.Debug("Alarm expression delete, metric id: " + rawMetricId)

		c.Status(http.StatusNoContent)
	}
}
