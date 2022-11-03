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
func UpdateSNMPHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// get container id
		uid, err := strconv.Atoi(c.Param("containerId"))
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}
		id := int32(uid)

		// bind container
		var container models.Container[models.SNMPContainer]
		err = c.ShouldBind(&container)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// validate container
		err = api.Validate.Struct(container)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// set type and container id
		container.Base.Type = types.CTSNMP
		container.Base.Id = id
		container.Protocol.ContainerId = id

		// get ident and target port existence
		ie, tpe, err := api.PgConn.SNMPContainers.AvailableIdentAndTargetPort(ctx,
			container.Base.Ident,
			container.Protocol.Target,
			container.Protocol.Port,
			id,
		)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to check container ident and target port existence", logger.ErrField(err))
			return
		}

		// check if ident exists
		if ie {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgIdentExists))
			return
		}

		// check if target port exists
		if tpe {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgTargetPortExists))
		}

		// create container
		e, err := api.PgConn.Containers.Update(ctx, container.Base)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to update container", logger.ErrField(err))
			return
		}

		// check if container exists
		if !e {
			c.Status(http.StatusNotFound)
			return
		}
		api.Log.Debug("container updated, ident: " + container.Base.Ident)

		// update snmp container
		e, err = api.PgConn.SNMPContainers.Update(ctx, container.Protocol)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to update snmp container", logger.ErrField(err))
			return
		}
		if !e {
			c.Status(http.StatusNotFound)
			api.Log.Error("base container exists but snmp container don't")
			return
		}
		api.Log.Debug("snmp container updated, target: " + container.Protocol.Target)

		c.Status(http.StatusOK)
	}
}
