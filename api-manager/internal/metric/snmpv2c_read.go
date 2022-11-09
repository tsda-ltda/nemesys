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

// Get a SNMPv2c metric.
// Responses:
//   - 400 If invalid params.
//   - 404 If not found.
//   - 200 If succeeded.
func GetSNMPv2cHandler(api *api.API, ct types.ContainerType) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// metric id
		metricId, err := strconv.ParseInt(c.Param("metricId"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		// get metric base information
		e, base, err := api.PgConn.Metrics.Get(ctx, metricId)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to get metric", logger.ErrField(err))
			return
		}

		// check if exists
		if !e {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgMetricNotFound))
			return
		}

		// get snmp metric
		e, snmp, err := api.PgConn.SNMPv2cMetrics.Get(ctx, metricId)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to get SNMP metric", logger.ErrField(err))
			return
		}

		// check if exists
		if !e {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgMetricNotFound))
			return
		}

		metric := models.Metric[models.SNMPMetric]{
			Base:     base,
			Protocol: snmp,
		}

		c.JSON(http.StatusOK, metric)
	}
}
