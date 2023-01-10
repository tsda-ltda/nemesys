package cost

import (
	"net/http"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/gin-gonic/gin"
)

// Get the server base plan.
// Responses:
//   - 200 If succeeded.
func GetBasePlanHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		basePlan, err := api.PG.GetBasePlan(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to get base plan on databae", logger.ErrField(err))
			return
		}

		c.JSON(http.StatusOK, tools.DataRes(basePlan))
	}
}

// Update the server base plan.
// Responses:
//   - 400 If invalid body.
//   - 200 If succeeded.
func UpdateBasePlanHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		var basePlan models.ServerBasePlan
		err := c.ShouldBind(&basePlan)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidBody))
			return
		}

		err = api.Validate.Struct(basePlan)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidJSONFields))
			return
		}

		err = api.PG.UpdateBasePlan(ctx, basePlan)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to update base plan", logger.ErrField(err))
			return
		}
		api.Log.Info("Base plan updated")

		c.JSON(http.StatusOK, tools.EmptyRes())
	}
}
