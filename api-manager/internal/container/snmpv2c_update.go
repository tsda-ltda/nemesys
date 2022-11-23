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
//   - 400 If target:port is in use.
//   - 200 If succeeded.
func UpdateSNMPv2cHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		id, err := strconv.ParseInt(c.Param("containerId"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		var container models.Container[models.SNMPv2cContainer]
		err = c.ShouldBind(&container)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidBody))
			return
		}

		err = api.Validate.Struct(container)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidJSONFields))
			return
		}

		container.Base.Id = int32(id)
		container.Protocol.Id = int32(id)
		container.Base.Type = types.CTSNMPv2c

		exists, err := api.PG.AvailableSNMPv2cContainerTargetPort(ctx,
			container.Protocol.Target,
			container.Protocol.Port,
			container.Base.Id,
		)
		if err != nil {
			api.Log.Error("fail to check if target port exists", logger.ErrField(err))
			c.Status(http.StatusInternalServerError)
			return
		}
		if exists {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgTargetPortExists))
		}

		exists, err = api.PG.UpdateSNMPv2cContainer(ctx, container)
		if err != nil {
			api.Log.Error("fail to update snmpv2c container", logger.ErrField(err))
			c.Status(http.StatusInternalServerError)
			return
		}
		if !exists {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgContainerNotFound))
			return
		}
		api.Log.Debug("snmp container updated, name: " + container.Base.Name)
		api.Amqph.NotifyContainerUpdated(container.Base, container.Protocol, types.CTSNMPv2c)

		c.Status(http.StatusOK)
	}
}
