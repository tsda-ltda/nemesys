package container

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/types"
	"github.com/gin-gonic/gin"
)

// Updates a SNMP container.
// Responses:
//   - 400 If invalid body.
//   - 400 If json fields are invalid.
//   - 400 If ident or target:port is in use.
//   - 200 If succeeded.
func UpdateSNMPv2cHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// get container id
		id, err := strconv.ParseInt(c.Param("containerId"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		// bind container
		var container models.Container[models.SNMPv2cContainer]
		err = c.ShouldBind(&container)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidBody))
			return
		}

		// validate container
		err = api.Validate.Struct(container)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidJSONFields))
			return
		}

		// assign values
		container.Base.Type = types.CTSNMPv2c
		container.Base.Id = int32(id)
		container.Protocol.ContainerId = int32(id)

		// get target port existence
		tpe, err := api.PgConn.SNMPv2cContainers.AvailableTargetPort(ctx,
			container.Protocol.Target,
			container.Protocol.Port,
			container.Base.Id,
		)
		if err != nil {
			api.Log.Error("fail to check if target port exists", logger.ErrField(err))
			c.Status(http.StatusInternalServerError)
			return
		}

		// check if target port exists
		if tpe {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgTargetPortExists))
		}

		// create container
		e, err := api.PgConn.Containers.Update(ctx, container.Base)
		if err != nil {
			api.Log.Error("fail to update container", logger.ErrField(err))
			c.Status(http.StatusInternalServerError)
			return
		}

		// check if container exists
		if !e {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgContainerNotFound))
			return
		}

		// update snmp container
		e, err = api.PgConn.SNMPv2cContainers.Update(ctx, container.Protocol)
		if err != nil {
			api.Log.Error("fail to update snmp container", logger.ErrField(err))
			c.Status(http.StatusInternalServerError)
			return
		}
		if !e {
			api.Log.Error("base container exists but snmp container don't")
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgContainerNotFound))
			return
		}
		api.Log.Debug("snmp container updated, name: " + container.Base.Name)

		// notify container
		api.Amqph.NotifyContainerUpdated(container.Base, container.Protocol, types.CTSNMPv2c)

		c.Status(http.StatusOK)
	}
}
