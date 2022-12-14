package metric

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	t "github.com/fernandotsda/nemesys/shared/amqph/tools"
	"github.com/fernandotsda/nemesys/shared/logger"

	"github.com/gin-gonic/gin"
)

// Delete a metric.
// Responses:
//   - 404 If not found.
//   - 200 If succeeded.
func DeleteHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		containerId, err := strconv.ParseInt(c.Param("containerId"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}

		rawId := c.Param("metricId")
		id, err := strconv.ParseInt(rawId, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}

		exists, err := api.PG.ContainerExist(ctx, int32(containerId))
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to check if container exists", logger.ErrField(err))
			return
		}
		if !exists {
			c.JSON(http.StatusNotFound, tools.MsgRes(tools.MsgContainerNotFound))
			return
		}

		exists, err = api.PG.DeleteMetric(ctx, id)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to delete metric", logger.ErrField(err))
			return
		}

		if !exists {
			c.JSON(http.StatusNotFound, tools.MsgRes(tools.MsgMetricNotFound))
			return
		}
		api.Log.Debug("Metric deleted, id: " + rawId)
		t.NotifyMetricDeleted(api.Amqph, int32(containerId), id)

		c.JSON(http.StatusOK, tools.EmptyRes())
	}
}
