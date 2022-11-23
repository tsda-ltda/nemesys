package datapolicy

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
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

		id, err := strconv.ParseInt(c.Param("id"), 10, 16)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		var dp models.DataPolicy
		err = c.ShouldBind(&dp)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidBody))
			return
		}

		err = api.Validate.Struct(dp)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidJSONFields))
			return
		}

		dp.Id = int16(id)
		tx, exists, err := api.PG.UpdateDataPolicy(ctx, dp)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to update data policy on influxdb", logger.ErrField(err))
			return
		}
		if !exists {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgDataPolicyNotFound))
			return
		}

		err = api.Influx.UpdateDataPolicy(ctx, dp)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to update data policy on influxdb", logger.ErrField(err))
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

		api.Log.Info("data policy updated, id: " + fmt.Sprint(id))

		c.Status(http.StatusOK)
	}
}
