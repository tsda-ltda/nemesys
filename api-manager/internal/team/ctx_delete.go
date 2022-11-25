package team

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
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

		rawId := c.Param("ctxId")
		id, err := strconv.ParseInt(rawId, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		exists, err := api.PG.DeleteContext(ctx, int32(id))
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			api.Log.Error("fail to delete context", logger.ErrField(err))
			c.Status(http.StatusInternalServerError)
			return
		}
		if !exists {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgContextNotFound))
			return
		}

		api.Log.Debug("context deleted, id: " + rawId)
		c.Status(http.StatusNoContent)
	}
}
