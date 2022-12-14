package category

import (
	"net/http"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/gin-gonic/gin"
)

// Create an alarm category.
// Responses:
//   - 400 If invalid body.
//   - 400 If category level already exists.
//   - 200 If succeeded.
func CreateHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		var category models.AlarmCategory
		err := c.ShouldBind(&category)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidBody))
			return
		}

		err = api.Validate.Struct(category)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidJSONFields))
			return
		}

		exists, err := api.PG.CategoryLevelExists(ctx, category.Level, -1)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail check if category level exists", logger.ErrField(err))
			return
		}
		if exists {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgAlarmCategoryLevelExists))
			return
		}

		id, err := api.PG.CreateAlarmCategory(ctx, category)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			api.Log.Error("Fail to create alarm category", logger.ErrField(err))
			c.Status(http.StatusInternalServerError)
			return
		}
		api.Log.Debug("Alarm category created with success, name: " + category.Name)

		c.JSON(http.StatusOK, tools.IdRes(int64(id)))
	}
}
