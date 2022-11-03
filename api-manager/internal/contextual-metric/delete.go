package ctxmetric

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
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

		// contextual metric id
		rawId := c.Param("metricId")
		id, err := strconv.Atoi(rawId)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// delete
		e, err := api.PgConn.ContextualMetrics.Delete(ctx, int64(id))
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to delete contextual metric", logger.ErrField(err))
			return
		}

		// check if exists
		if !e {
			c.Status(http.StatusNotFound)
			return
		}
		api.Log.Debug("contextual metric deleted, id: " + rawId)

		c.Status(http.StatusNoContent)
	}
}
