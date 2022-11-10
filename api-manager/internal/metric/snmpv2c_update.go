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

// Updates a SNMPv2c metric.
// Responses:
//   - 400 If invalid body.
//   - 400 If json fields are invalid.
//   - 404 If container or metric not found.
//   - 200 If succeeded.
func UpdateSNMPv2cHandler(api *api.API, ct types.ContainerType) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// get container id
		rawContainerId := c.Param("containerId")
		containerId, err := strconv.ParseInt(rawContainerId, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		// get metric id
		rawId := c.Param("metricId")
		id, err := strconv.ParseInt(rawId, 10, 64)
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

		// assign id
		metric.Base.Id = id
		metric.Protocol.Id = id
		metric.Base.ContainerId = int32(containerId)

		// get if container and data policy exists
		e, ce, dpe, err := api.PgConn.Metrics.ExistsContainerAndDataPolicy(ctx, metric.Base.ContainerId, types.CTSNMPv2c, metric.Base.DataPolicyId, id)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to check container and data policy existence", logger.ErrField(err))
			return
		}

		// check metric if exists
		if !e {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgMetricNotFound))
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

		// create SNMP metric
		e, err = api.PgConn.SNMPv2cMetrics.Update(ctx, metric.Protocol)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to update snmp metric", logger.ErrField(err))
			return
		}

		// check if exists
		if !e {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgMetricNotFound))
			return
		}

		// update metric
		e, err = api.PgConn.Metrics.Update(ctx, metric.Base)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to update base update", logger.ErrField(err))
			return
		}

		// check if metric exists
		if !e {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgMetricNotFound))
			api.Log.Error("snmp metric exist but base metric don't")
			return
		}
		api.Log.Debug("metric updated, name" + metric.Base.Name)
		api.Amqph.NotifyMetricUpdated(metric.Base, metric.Protocol, types.CTSNMPv2c)

		c.Status(http.StatusOK)
	}
}
