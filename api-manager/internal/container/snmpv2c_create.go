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
//   - 400 If target:port is in use.
//   - 200 If succeeded.
func CreateSNMPv2cHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		var container models.Container[models.SNMPv2cContainer]
		err := c.ShouldBind(&container)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidBody))
			return
		}

		err = api.Validate.Struct(container)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidJSONFields))
			return
		}

		container.Base.Type = types.CTSNMPv2c

		exists, err := api.PG.AvailableSNMPv2cContainerTargetPort(ctx,
			container.Protocol.Target,
			container.Protocol.Port,
			-1,
		)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to check if target port exists", logger.ErrField(err))
			return
		}
		if exists {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgTargetPortExists))
			return
		}

		err = api.PG.CreateSNMPv2cContainer(ctx, container)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to crate snmpv2c container", logger.ErrField(err))
			return
		}
		api.Log.Debug("snmp container crated, name: " + container.Base.Name)
		api.Amqph.NotifyContainerCreated(container.Base, container.Protocol, types.CTSNMPv2c)

		c.Status(http.StatusOK)
	}
}
