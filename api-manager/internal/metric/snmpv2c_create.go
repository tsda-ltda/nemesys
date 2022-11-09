package metric

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

// Creates a SNMP metric.
// Responses:
//   - 400 If invalid body.
//   - 400 If json fields are invalid.
//   - 404 If container not found.
//   - 200 If succeeded.
func CreateSNMPHandler(api *api.API, ct types.ContainerType) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// container id
		containerId, err := strconv.ParseInt(c.Param("containerId"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		// bind body
		var metric models.Metric[models.SNMPMetric]
		err = c.ShouldBind(&metric)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidBody))
			return
		}

		// validate body
		err = api.Validate.Struct(metric)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidJSONFields))
			return
		}

		// assign containerId and container type
		metric.Base.ContainerId = int32(containerId)
		metric.Base.ContainerType = ct

		// get if container and data policy exists
		_, ce, dpe, err := api.PgConn.Metrics.ExistsContainerAndDataPolicy(ctx, metric.Base.ContainerId, ct, metric.Base.DataPolicyId, -1)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to check container and data policy existence", logger.ErrField(err))
			return
		}

		// check if data policy exists
		if !dpe {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgDataPolicyNotFound))
			return
		}

		// check if container exists
		if !ce {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgContainerNotFound))
			return
		}

		id, err := api.PgConn.Metrics.Create(ctx, metric.Base)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to create base metric", logger.ErrField(err))
			return
		}
		api.Log.Debug("base metric created, name: " + metric.Base.Name)

		// assign id
		metric.Protocol.Id = id

		switch ct {
		case types.CTSNMPv2c:
			err = api.PgConn.SNMPv2cMetrics.Create(ctx, metric.Protocol)
			if err != nil {
				c.Status(http.StatusInternalServerError)
				api.Log.Error("fail to create snmp metric", logger.ErrField(err))
				return
			}
			api.Log.Debug("snmp metric created, name" + metric.Base.Name)
		default:
			api.Log.Error("fail to create metric, unsupported container type")
			c.Status(http.StatusInternalServerError)
			return
		}

		c.Status(http.StatusOK)
	}
}