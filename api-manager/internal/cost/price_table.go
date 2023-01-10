package cost

import (
	"net/http"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/gin-gonic/gin"
)

// Get the server price table.
// Responses:
//   - 200 If succeeded.
func GetPriceTableHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		table, err := api.PG.GetPriceTable(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to get price table on databae", logger.ErrField(err))
			return
		}

		c.JSON(http.StatusOK, tools.DataRes(table))
	}
}

// Update the server price table.
// Responses:
//   - 400 If invalid body.
//   - 200 If succeeded.
func UpdatePriceTableHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		var table models.ServerPriceTable
		err := c.ShouldBind(&table)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidBody))
			return
		}

		err = api.Validate.Struct(table)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidJSONFields))
			return
		}

		err = api.PG.UpdatePriceTable(ctx, table)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to update price table", logger.ErrField(err))
			return
		}
		api.Log.Info("Price table updated")

		c.JSON(http.StatusOK, tools.EmptyRes())
	}
}
