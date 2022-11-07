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

// Updates a SNMP metric.
// Responses:
//   - 400 If invalid body.
//   - 400 If json fields are invalid.
//   - 400 If ident is in use.
//   - 404 If container or metric not found.
//   - 200 If succeeded.
func UpdateSNMPHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// get container id
		rawContainerId := c.Param("containerId")
		containerId, err := strconv.ParseInt(rawContainerId, 10, 0)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// get metric id
		rawId := c.Param("metricId")
		id, err := strconv.ParseInt(rawId, 10, 0)
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

		// assign id
		metric.Base.Id = id
		metric.Protocol.Id = id
		metric.Base.ContainerId = int32(containerId)

		// get if ident, container and data policy exists
		e, ce, dpe, ie, err := api.PgConn.Metrics.ExistsIdentAndContainerAndDataPolicy(ctx, metric.Base.ContainerId, types.CTSNMP, metric.Base.DataPolicyId, metric.Base.Ident, id)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to check metric's ident, container and data policy existence", logger.ErrField(err))
			return
		}

		// check metric if exists
		if !e {
			c.Status(http.StatusNotFound)
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

		// create SNMP metric
		e, err = api.PgConn.SNMPMetrics.Update(ctx, metric.Protocol)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to update snmp metric", logger.ErrField(err))
			return
		}

		// check if exists
		if !e {
			c.Status(http.StatusNotFound)
			return
		}
		api.Log.Debug("snmp metric update, base metric id: " + strconv.FormatInt(id, 10))

		// update metric
		e, err = api.PgConn.Metrics.Update(ctx, metric.Base)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to update base update", logger.ErrField(err))
			return
		}

		// check if metric exists
		if !e {
			c.Status(http.StatusNotFound)
			api.Log.Error("snmp metric exist but base metric don't")
			return
		}
		api.Log.Debug("base metric updated, ident" + metric.Base.Ident)

		c.Status(http.StatusOK)
	}
}
