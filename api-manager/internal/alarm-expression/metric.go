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

// Creates a metric relation with alarm expression.
// Responses:
//   - 400 If invalid params.
//   - 404 If alarm expression not found.
//   - 404 If metric not found.
//   - 400 If relation already exists.
//   - 200 If succeeded.
func CreateMetricRelationHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		id, err := strconv.ParseInt(c.Param("expressionId"), 0, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		var id64 models.Id64
		err = c.ShouldBind(&id64)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidBody))
			return
		}

		err = api.Validate.Struct(id64)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidJSONFields))
			return
		}

		r, err := api.PG.MetricAlarmExpressionRelExists(ctx, int32(id), id64.Id)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to get alarm expression, metric and relation existence", logger.ErrField(err))
			return
		}
		if !r.Exists {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgAlarmExpressionNotFound))
			return
		}
		if !r.MetricExists {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgMetricNotFound))
			return
		}
		if r.RelationExists {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgAlarmExpressionAndMetricRelExists))
			return
		}

		err = api.PG.CrateMetricAlarmExpressionRel(ctx, int32(id), id64.Id)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to create alarm expression and metric relation", logger.ErrField(err))
			return
		}
		api.Log.Debug("Metric and alarm expression created, metric id: " + strconv.FormatInt(id64.Id, 10))

		c.Status(http.StatusOK)
	}
}

// Remove a metric relation with alarm expression.
// Responses:
//   - 400 If invalid params.
//   - 404 If relation already not found.
//   - 200 If succeeded.
func DeleteMetricRelationHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		expressionId, err := strconv.ParseInt(c.Param("expressionId"), 0, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		rawMetricId := c.Param("metricId")
		metricId, err := strconv.ParseInt(rawMetricId, 0, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}
		exists, err := api.PG.RemoveMetricAlarmExpressionRel(ctx, int32(expressionId), metricId)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to remove alarm expression and metric relation", logger.ErrField(err))
			return
		}
		if !exists {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgAlarmExpressionAndMetricRelNotFound))
			return
		}

		api.Log.Debug("Metric and alarm expression removed, metric id: " + rawMetricId)

		c.Status(http.StatusOK)
	}
}
