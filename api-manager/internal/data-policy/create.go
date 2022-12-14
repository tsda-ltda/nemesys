package datapolicy

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/env"
	"github.com/fernandotsda/nemesys/shared/influxdb"
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
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}

		err = api.Validate.Struct(dp)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidBody))
			return
		}

		if !influxdb.ValidateAggrFunction(dp.AggrFn) {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidAggrFn))
			return
		}

		n, err := api.PG.CountDataPolicy(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to count number of data polcies", logger.ErrField(err))
			return
		}

		max, err := strconv.ParseInt(env.MaxDataPolicies, 10, 0)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to parse env.MaxDataPolicies", logger.ErrField(err))
			return
		}
		if n >= max {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgMaxDataPolicy))
			api.Log.Warn("Attempt to create data-policy failed, maximum number reached")
			return
		}

		tx, id, err := api.PG.CreateDataPolicy(ctx, dp)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			api.Log.Error("Fail to create data policy on postgres", logger.ErrField(err))
			c.Status(http.StatusInternalServerError)
			return
		}

		dp.Id = id

		err = api.Influx.CreateDataPolicy(ctx, dp)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			api.Log.Error("Fail to create data policy on influxdb", logger.ErrField(err))
			c.Status(http.StatusInternalServerError)
			err = tx.Rollback()
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				api.Log.Error("Fail to rollback tx", logger.ErrField(err))
				return
			}
			return
		}
		err = tx.Commit()
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to commit tx", logger.ErrField(err))
			return
		}

		api.Log.Info("Data policy created, id: " + strconv.FormatInt(int64(id), 10))
		c.JSON(http.StatusOK, tools.IdRes(int64(id)))
	}
}
