package team

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/gin-gonic/gin"
)

// Deletes a context.
// Responses:
//   - 404 If context not found.
//   - 204 If succeeded.
func DeleteContextHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// get context id
		rawId := c.Param("contextId")
		id, err := strconv.Atoi(rawId)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// delete context
		e, err := api.PgConn.Contexts.Delete(ctx, id)
		if err != nil {
			api.Log.Error("fail to delete context", logger.ErrField(err))
			c.Status(http.StatusInternalServerError)
			return
		}

		// check if context existed
		if !e {
			c.Status(http.StatusNotFound)
			return
		}

		api.Log.Debug("context deleted, id: " + rawId)
		c.Status(http.StatusNoContent)
	}
}
