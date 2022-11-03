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
//   - 400 If ident is in use.
//   - 404 If container not found.
//   - 200 If succeeded.
func CreateSNMPHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// container id
		containerId, err := strconv.Atoi(c.Param("containerId"))
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// bind body
		var metric models.Metric[models.SNMPMetric]
		err = c.ShouldBind(&metric)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// validate body
		err = api.Validate.Struct(metric)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// assign containerId and container type
		metric.Base.ContainerId = int32(containerId)
		metric.Base.ContainerType = types.CTSNMP

		// get if ident, container and data policy exists
		_, ce, dpe, ie, err := api.PgConn.Metrics.ExistsIdentAndContainerAndDataPolicy(ctx, metric.Base.ContainerId, types.CTSNMP, metric.Base.DataPolicyId, metric.Base.Ident, -1)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to check metric's ident, container and data policy existence", logger.ErrField(err))
			return
		}

		// check if data policy exists
		if !dpe {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgDataPolicyNotFound))
			return
		}

		// check if ident already in use
		if ie {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgIdentExists))
			return
		}

		// check if container exists
		if !ce {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgContainerNotFound))
			return
		}

		// create metric
		id, err := api.PgConn.Metrics.Create(ctx, metric.Base)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to create base metric", logger.ErrField(err))
			return
		}
		api.Log.Debug("base metric created, ident: " + metric.Base.Ident)

		// assign id
		metric.Protocol.MetricId = id

		// create SNMP metric
		err = api.PgConn.SNMPMetrics.Create(ctx, metric.Protocol)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to create snmp metric", logger.ErrField(err))
			return
		}
		api.Log.Debug("snmp metric created, base metric id: " + strconv.FormatInt(id, 10))

		c.Status(http.StatusOK)
	}
}
