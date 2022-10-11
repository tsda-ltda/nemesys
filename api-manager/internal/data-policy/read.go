package datapolicy

import (
	"net/http"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/gin-gonic/gin"
)

// Get all data policies.
// Responses:
//   - 200 If succeeded.
func MGetHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// get data policies
		sql := `SELECT id, descr, retention, use_aggregation, aggregation_retention, aggregation_interval
		FROM data_policies;`
		rows, err := api.PgConn.Query(ctx, sql)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to read data policies", logger.ErrField(err))
			return
		}

		// scan rows
		dataPolicies := []models.DataPolicy{}
		for rows.Next() {
			var dp models.DataPolicy
			rows.Scan(
				&dp.Id,
				&dp.Descr,
				&dp.Retention,
				&dp.UseAggregation,
				&dp.AggregationRetention,
				&dp.AggregationInterval,
			)
			dataPolicies = append(dataPolicies, dp)
		}

		c.JSON(http.StatusOK, dataPolicies)
	}
}
