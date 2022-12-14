package container

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	t "github.com/fernandotsda/nemesys/shared/amqph/tools"
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
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}

		var container models.Container[models.SNMPv2cContainer]
		err = c.ShouldBind(&container)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidBody))
			return
		}

		err = api.Validate.Struct(container)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidJSONFields))
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
			if ctx.Err() != nil {
				return
			}
			api.Log.Error("Fail to check if target port exists", logger.ErrField(err))
			c.Status(http.StatusInternalServerError)
			return
		}
		if exists {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgTargetPortExists))
		}

		exists, err = api.PG.UpdateSNMPv2cContainer(ctx, container)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			api.Log.Error("Fail to update SNMPv2c container", logger.ErrField(err))
			c.Status(http.StatusInternalServerError)
			return
		}
		if !exists {
			c.JSON(http.StatusNotFound, tools.MsgRes(tools.MsgContainerNotFound))
			return
		}
		api.Log.Debug("SNMPv2c container updated, name: " + container.Base.Name)
		t.NotifyContainerUpdated(api.Amqph, container.Base, container.Protocol)

		c.JSON(http.StatusOK, tools.EmptyRes())
	}
}
