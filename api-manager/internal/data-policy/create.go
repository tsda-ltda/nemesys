package datapolicy

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/shared/env"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/gin-gonic/gin"
)

// Creates a new data policy.
// Responses:
//   - 400 If invalid body.
//   - 400 If json fields are invalid.
//   - 400 If exceeds the maximum number of data policies.
//   - 200 If succeeded.
func CreateHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// bind data policy
		var dp models.DataPolicy
		err := c.ShouldBind(&dp)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// validate struct
		err = api.Validate.Struct(dp)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// count number of data policies already created
		sql := `SELECT COUNT(*) FROM data_policies;`
		var n int
		err = api.PgConn.QueryRow(ctx, sql).Scan(&n)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to count number of data polcies", logger.ErrField(err))
			return
		}

		// check if number of data policies is the maximum number permited
		max, err := strconv.Atoi(env.MaxDataPolicies)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to parse env.MaxDataPolicies", logger.ErrField(err))
			return
		}
		if n >= max {
			c.Status(http.StatusBadRequest)
			api.Log.Info("attempt to create data-policy failed, maximum number reached")
			return
		}

		// create data policy
		sql = `INSERT INTO data_policies (descr, use_aggregation, retention, aggregation_retention, aggregation_interval)
		VALUES ($1, $2, $3, $4, $5);`
		_, err = api.PgConn.Exec(ctx, sql, dp.Descr, dp.UseAggregation, dp.Retention, dp.AggregationRetention, dp.AggregationInterval)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to create new data policy", logger.ErrField(err))
			return
		}
		api.Log.Info("new data policy added")

		c.Status(http.StatusOK)
	}
}
