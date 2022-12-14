package datapolicy

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/influxdb"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/gin-gonic/gin"
)

// Update a data policy.
// Responses:
//   - 400 If invalid id.
//   - 400 If invalid body.
//   - 400 If invalid body fields.
//   - 404 If data policy not found.
//   - 200 If succeeded.
func UpdateHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		id, err := strconv.ParseInt(c.Param("dpId"), 10, 16)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}

		var dp models.DataPolicy
		err = c.ShouldBind(&dp)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidBody))
			return
		}

		err = api.Validate.Struct(dp)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidJSONFields))
			return
		}

		if !influxdb.ValidateAggrFunction(dp.AggrFn) {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidAggrFn))
			return
		}

		dp.Id = int16(id)
		tx, exists, err := api.PG.UpdateDataPolicy(ctx, dp)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to update data policy on influxdb", logger.ErrField(err))
			return
		}
		if !exists {
			c.JSON(http.StatusNotFound, tools.MsgRes(tools.MsgDataPolicyNotFound))
			return
		}

		err = api.Influx.UpdateDataPolicy(ctx, dp)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to update data policy on influxdb", logger.ErrField(err))
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

		api.Log.Info("Data policy updated, id: " + strconv.FormatInt(int64(id), 10))

		c.JSON(http.StatusOK, tools.EmptyRes())
	}
}
