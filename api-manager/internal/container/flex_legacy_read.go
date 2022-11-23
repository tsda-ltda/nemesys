package container

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/gin-gonic/gin"
)

// Get a Flex Legacy container.
// Responses:
//   - 404 If not found.
//   - 200 If succeeded.
func GetFlexLegacyHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		id, err := strconv.ParseInt(c.Param("containerId"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		r, err := api.PG.GetFlexLegacyContainer(ctx, int32(id))
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to get flex legacy container", logger.ErrField(err))
			return
		}
		if !r.Exists {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgContainerNotFound))
			return
		}

		c.JSON(http.StatusOK, r.Container)
	}
}
