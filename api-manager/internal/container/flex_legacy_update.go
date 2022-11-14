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

		// get container id
		id, err := strconv.ParseInt(c.Param("containerId"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		// bind body
		var container models.Container[models.FlexLegacyContainer]
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
		container.Base.Id = int32(id)
		container.Protocol.Id = int32(id)

		// check if container, target port combination and serial number exists
		r, err := api.PgConn.FlexLegacyContainers.ExistsContainerTargetPortAndSerialNumber(ctx,
			int32(id),
			container.Protocol.Target,
			container.Protocol.Port,
			container.Protocol.SerialNumber,
		)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to check if container, target port combination and serial-number exists", logger.ErrField(err))
			return
		}

		// check if container exists
		if !r.ContainerExists {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgContainerNotFound))
			return
		}

		// check if target port combination exists
		if r.TargetPortExists {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgTargetPortExists))
			return
		}

		// check if serial-number exists
		if r.SerialNumberExists {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgSerialNumberExists))
			return
		}

		// update base container
		exists, err := api.PgConn.Containers.Update(ctx, container.Base)
		if err != nil {
			c.Status(http.StatusBadRequest)
			api.Log.Error("fail to create base container", logger.ErrField(err))
			return
		}

		// check if exists
		if !exists {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgContainerNotFound))
			return
		}

		// update flex legacy container
		exists, err = api.PgConn.FlexLegacyContainers.Update(ctx, container.Protocol)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to create flex legacy container", logger.ErrField(err))
			return
		}

		// check if exists
		if !exists {
			api.Log.Error("base container exists but flex-legacy container don't")
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgContainerNotFound))
			return
		}
		api.Log.Debug("flex legacy container updated with success, name: " + container.Base.Name)
		api.Amqph.NotifyContainerUpdated(container.Base, container.Protocol, types.CTFlexLegacy)

		c.Status(http.StatusOK)
	}
}
