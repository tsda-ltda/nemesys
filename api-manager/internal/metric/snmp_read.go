package metric

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/gin-gonic/gin"
)

// Get a SNMP metric.
// Responses:
//   - 400 If invalid params.
//   - 404 If not found.
//   - 200 If succeeded.
func GetSNMPHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// metric id
		metricId, err := strconv.ParseInt(c.Param("metricId"), 10, 0)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// get container base information
		e, base, err := api.PgConn.Metrics.Get(ctx, metricId)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to get container", logger.ErrField(err))
			return
		}

		// check if exists
		if !e {
			c.Status(http.StatusNotFound)
			return
		}

		// get snmp container
		e, snmp, err := api.PgConn.SNMPMetrics.Get(ctx, metricId)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to get SNMP metric", logger.ErrField(err))
			return
		}

		// check if exists
		if !e {
			c.Status(http.StatusNotFound)
			return
		}

		metric := models.Metric[models.SNMPMetric]{
			Base:     base,
			Protocol: snmp,
		}

		c.JSON(http.StatusOK, metric)
	}
}
