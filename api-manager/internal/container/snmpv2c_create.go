package container

import (
	"net/http"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	t "github.com/fernandotsda/nemesys/shared/amqph/tools"
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
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidBody))
			return
		}

		err = api.Validate.Struct(container)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidJSONFields))
			return
		}

		container.Base.Type = types.CTSNMPv2c

		exists, err := api.PG.AvailableSNMPv2cContainerTargetPort(ctx,
			container.Protocol.Target,
			container.Protocol.Port,
			-1,
		)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to check if target port exists", logger.ErrField(err))
			return
		}
		if exists {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgTargetPortExists))
			return
		}

		id, err := api.PG.CreateSNMPv2cContainer(ctx, container)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to crate SNMPv2c container", logger.ErrField(err))
			return
		}
		container.Base.Id = id
		container.Protocol.Id = id
		api.Log.Debug("SNMPv2c container crated, name: " + container.Base.Name)
		t.NotifyContainerCreated(api.Amqph, container.Base, container.Protocol)

		c.JSON(http.StatusOK, tools.IdRes(int64(id)))
	}
}
