package datapolicy

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	t "github.com/fernandotsda/nemesys/shared/amqph/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/gin-gonic/gin"
)

// Deletes a data policy.
// Responses:
//   - 400 If invalid id.
//   - 404 If data policy not found.
//   - 200 If succeeded.
func DeleteHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		id, err := strconv.ParseInt(c.Param("dpId"), 10, 16)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}

		exists, err := api.PG.DeleteDataPolicy(ctx, int16(id))
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to delete data policy from postgres", logger.ErrField(err))
			return
		}
		if !exists {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgDataPolicyNotFound))
			return
		}

		err = api.Influx.DeleteDataPolicy(ctx, int16(id))
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to delete data policy from influxdb", logger.ErrField(err))
			return
		}
		t.NotifyDataPolicyDeleted(api.Amqph, int16(id))

		api.Log.Info("Data policy deleted, id: " + strconv.FormatInt(int64(id), 10))
		c.JSON(http.StatusOK, tools.EmptyRes())
	}
}
