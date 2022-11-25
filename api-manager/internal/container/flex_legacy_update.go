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

// Updates a Flex Legacy container.
// Responses:
//   - 400 If invalid body.
//   - 400 If json fields are invalid.
//   - 400 If serial-number or target:port is in use.
//   - 404 If container not found.
//   - 200 If succeeded.
func UpdateFlexLegacy(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		id, err := strconv.ParseInt(c.Param("containerId"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		var container models.Container[models.FlexLegacyContainer]
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
		container.Base.Type = types.CTFlexLegacy

		r, err := api.PG.ExistsFlexLegacyContainerTargetPortAndSerialNumber(ctx,
			int32(id),
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
		if !r.ContainerExists {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgContainerNotFound))
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

		exists, err := api.PG.UpdateFlexLegacyContainer(ctx, container)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to update flex legacy container", logger.ErrField(err))
			return
		}

		if !exists {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgContainerNotFound))
			return
		}
		api.Log.Debug("flex legacy container updated with success, name: " + container.Base.Name)
		api.Amqph.NotifyContainerUpdated(container.Base, container.Protocol, types.CTFlexLegacy)

		c.Status(http.StatusOK)
	}
}
