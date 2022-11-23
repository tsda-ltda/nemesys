package datapolicy

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
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

		var dp models.DataPolicy
		err := c.ShouldBind(&dp)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		err = api.Validate.Struct(dp)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidBody))
			return
		}

		n, err := api.PG.CountDataPolicy(ctx)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to count number of data polcies", logger.ErrField(err))
			return
		}

		max, err := strconv.ParseInt(env.MaxDataPolicies, 10, 0)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to parse env.MaxDataPolicies", logger.ErrField(err))
			return
		}
		if n >= max {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgMaxDataPolicy))
			api.Log.Warn("attempt to create data-policy failed, maximum number reached")
			return
		}

		tx, id, err := api.PG.CreateDataPolicy(ctx, dp)
		if err != nil {
			api.Log.Error("fail to create data policy on postgres", logger.ErrField(err))
			c.Status(http.StatusInternalServerError)
			return
		}

		dp.Id = id

		err = api.Influx.CreateDataPolicy(ctx, dp)
		if err != nil {
			api.Log.Error("fail to create data policy on influxdb", logger.ErrField(err))
			c.Status(http.StatusInternalServerError)
			err = tx.Rollback()
			if err != nil {
				api.Log.Error("fail to rollback tx", logger.ErrField(err))
				return
			}
			return
		}
		err = tx.Commit()
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to commit tx", logger.ErrField(err))
			return
		}

		api.Log.Info("data policy created")
		c.Status(http.StatusOK)
	}
}
