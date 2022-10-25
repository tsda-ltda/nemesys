package metric

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/gin-gonic/gin"
)

// Delete a container and dependencies.
// Responses:
//   - 400 If invalid params.
//   - 404 If not found.
//   - 204 If succeeded.
func DeleteHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// get metric id
		rawId := c.Param("id")
		id, err := strconv.Atoi(rawId)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// delete metric
		e, err := api.PgConn.Metrics.Delete(ctx, id)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to delete metric", logger.ErrField(err))
			return
		}

		// check if metric existed
		if !e {
			c.Status(http.StatusNotFound)
			return
		}

		api.Log.Debug("metric deleted, id: " + rawId)
		c.Status(http.StatusNoContent)
	}
}
