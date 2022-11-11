package container

import (
	"net/http"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/types"
	"github.com/gin-gonic/gin"
)

// Creates a SNMP container.
// Responses:
//   - 400 If invalid body.
//   - 400 If json fields are invalid.
//   - 400 If ident or target:port is in use.
//   - 200 If succeeded.
func CreateSNMPv2cHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// bind container
		var container models.Container[models.SNMPv2cContainer]
		err := c.ShouldBind(&container)
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

		// set type
		container.Base.Type = types.CTSNMPv2c

		// get target port existence
		tpe, err := api.PgConn.SNMPv2cContainers.AvailableTargetPort(ctx,
			container.Protocol.Target,
			container.Protocol.Port,
			-1,
		)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to check if target port exists", logger.ErrField(err))
			return
		}

		// check if target port exists
		if tpe {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgTargetPortExists))
			return
		}

		// create container
		id, err := api.PgConn.Containers.Create(ctx, container.Base)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to crate container", logger.ErrField(err))
			return
		}

		// assign id
		container.Protocol.Id = int32(id)

		// create snmp container
		err = api.PgConn.SNMPv2cContainers.Create(ctx, container.Protocol)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to crate snmp container", logger.ErrField(err))
			return
		}
		api.Log.Debug("snmp container crated, name: " + container.Base.Name)

		// notify container
		api.Amqph.NotifyContainerCreated(container.Base, container.Protocol, types.CTSNMPv2c)

		c.Status(http.StatusOK)
	}
}
