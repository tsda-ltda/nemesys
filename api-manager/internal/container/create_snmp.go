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
func CreateSNMPHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// bind container
		var container models.Container[models.SNMPContainer]
		err := c.ShouldBind(&container)
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

		// set type
		container.Base.Type = types.CTSNMP

		// get ident and target port existence
		ie, tpe, err := api.PgConn.SNMPContainers.AvailableIdentAndTargetPort(ctx,
			container.Base.Ident,
			container.Protocol.Target,
			container.Protocol.Port,
			-1,
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
			return
		}

		// create container
		id, err := api.PgConn.Containers.Create(ctx, container.Base)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to crate container", logger.ErrField(err))
			return
		}
		api.Log.Debug("container created, ident: " + container.Base.Ident)

		// assign id
		container.Protocol.ContainerId = id

		// create snmp container
		err = api.PgConn.SNMPContainers.Create(ctx, container.Protocol)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to crate snmp container", logger.ErrField(err))
			return
		}
		api.Log.Debug("snmp container crated, target: " + container.Protocol.Target)

		c.Status(http.StatusOK)
	}
}
