package container

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	t "github.com/fernandotsda/nemesys/shared/amqph/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/gin-gonic/gin"
)

// Delete a container and dependencies.
// Responses:
//   - 400 If invalid params.
//   - 404 If not found.
//   - 200 If succeeded.
func DeleteHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// get container id
		rawId := c.Param("containerId")
		id, err := strconv.ParseInt(rawId, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}

		// delete container
		e, err := api.PG.DeleteContainer(ctx, int32(id))
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to delete container", logger.ErrField(err))
			return
		}

		// check if exists
		if !e {
			c.JSON(http.StatusNotFound, tools.MsgRes(tools.MsgContainerNotFound))
			return
		}
		api.Log.Debug("Container deleted, id: " + rawId)
		t.NotifyContainerDeleted(api.Amqph, int32(id))

		c.JSON(http.StatusOK, tools.EmptyRes())
	}
}
