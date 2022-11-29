package ctxmetric

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/gin-gonic/gin"
)

// Deletes a contextual metric.
// Responses:
//   - 400 If invalid id.
//   - 404 If contextual metric does not exists.
//   - 204 If succeeded.
func DeleteHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		rawId := c.Param("metricId")
		id, err := strconv.ParseInt(rawId, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		e, err := api.PG.DeleteContextualMetric(ctx, int64(id))
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to delete contextual metric", logger.ErrField(err))
			return
		}

		if !e {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgContextualMetricNotFound))
			return
		}
		api.Log.Debug("Contextual metric deleted, id: " + rawId)

		c.Status(http.StatusNoContent)
	}
}
