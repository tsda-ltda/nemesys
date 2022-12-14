package container

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/gin-gonic/gin"
)

// Get a SNMP container.
// Responses:
//   - 404 If not found.
//   - 200 If succeeded.
func GetSNMPv2cHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		id, err := strconv.ParseInt(c.Param("containerId"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}

		exists, container, err := api.PG.GetSNMPv2cContainer(ctx, int32(id))
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to get SNMPv2c container")
			return
		}
		if !exists {
			c.JSON(http.StatusNotFound, tools.MsgRes(tools.MsgContainerNotFound))
			return
		}

		c.JSON(http.StatusOK, tools.DataRes(container))
	}
}
