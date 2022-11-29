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

// Creates a Flex Legacy container.
// Responses:
//   - 400 If invalid body.
//   - 400 If json fields are invalid.
//   - 400 If serial-number or target:port is in use.
//   - 200 If succeeded.
func CreateFlexLegacy(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		var container models.Container[models.FlexLegacyContainer]
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

		container.Base.Type = types.CTFlexLegacy

		r, err := api.PG.ExistsFlexLegacyContainerTargetPortAndSerialNumber(ctx,
			-1,
			container.Protocol.Target,
			container.Protocol.Port,
			container.Protocol.SerialNumber,
		)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to check if container, target port combination and serial-number exists", logger.ErrField(err))
			return
		}
		if r.TargetPortExists {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgTargetPortExists))
			return
		}
		if r.SerialNumberExists {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgSerialNumberExists))
			return
		}

		id, err := api.PG.CreateFlexLegacyContainer(ctx, container)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to create flex legacy container", logger.ErrField(err))
			return
		}
		container.Base.Id = id
		container.Protocol.Id = id
		api.Log.Debug("flex legacy container created with success, name: " + container.Base.Name)
		api.Amqph.NotifyContainerCreated(container.Base, container.Protocol, types.CTFlexLegacy)

		c.Status(http.StatusOK)
	}
}
