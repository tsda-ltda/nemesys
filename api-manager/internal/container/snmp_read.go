package container

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/gin-gonic/gin"
)

// Get a SNMP container.
// Responses:
//   - 404 If not found.
//   - 200 If succeeded.
func GetSNMPHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// get container id
		id, err := strconv.ParseInt(c.Param("containerId"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		// get container base information
		e, base, err := api.PgConn.Containers.Get(ctx, int32(id))
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to get container", logger.ErrField(err))
			return
		}

		// check if exists
		if !e {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgContainerNotFound))
			return
		}

		// get snmp container
		e, snmp, err := api.PgConn.SNMPContainers.Get(ctx, int32(id))
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to get SNMP container", logger.ErrField(err))
			return
		}

		// check if exists
		if !e {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgContainerNotFound))
			return
		}

		container := models.Container[models.SNMPContainer]{
			Base:     base,
			Protocol: snmp,
		}

		c.JSON(http.StatusOK, container)
	}
}
