package alarmexp

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/gin-gonic/gin"
)

// Crates a alarm expression.
// Responses:
//   - 400 If invalid params.
//   - 404 If metric not found.
//   - 400 If alarm expression already exists.
//   - 200 If succeeded.
func CreateHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		rawMetricId := c.Param("metricId")
		metricId, err := strconv.ParseInt(rawMetricId, 0, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		var alarmExp models.AlarmExpression
		err = c.ShouldBind(&alarmExp)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidBody))
			return
		}

		err = api.Validate.Struct(alarmExp)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidJSONFields))
			return
		}

		expressionExists, metricExists, err := api.PG.AlarmExpressionExists(ctx, metricId)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to check if alarm expression exists", logger.ErrField(err))
			return
		}
		if !metricExists {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgMetricNotFound))
			return
		}
		if expressionExists {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgAlarmExpressionExists))
			return
		}

		alarmExp.MetricId = metricId
		err = api.PG.CreateAlarmExpression(ctx, alarmExp)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to create alarm expression", logger.ErrField(err))
			return
		}
		api.Log.Debug("Alarm expression created, metric id: " + rawMetricId)

		c.Status(http.StatusOK)
	}
}
