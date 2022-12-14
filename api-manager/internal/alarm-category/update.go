package category

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/gin-gonic/gin"
)

// Update an alarm category.
// Responses:
//   - 400 If invalid param.
//   - 400 If invalid body.
//   - 404 If not found.
//   - 400 If category level already exists.
//   - 200 If succeeded.
func UpdateHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		id, err := strconv.ParseInt(c.Param("categoryId"), 0, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}

		var category models.AlarmCategory
		err = c.ShouldBind(&category)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidBody))
			return
		}

		err = api.Validate.Struct(category)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidJSONFields))
			return
		}
		category.Id = int32(id)

		exists, err := api.PG.CategoryLevelExists(ctx, category.Level, category.Id)
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

		exists, err = api.PG.UpdateAlarmCategory(ctx, category)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			api.Log.Error("Fail to update alarm category", logger.ErrField(err))
			c.Status(http.StatusInternalServerError)
			return
		}
		if !exists {
			c.JSON(http.StatusNotFound, tools.MsgRes(tools.MsgAlarmCategoryNotFound))
			return
		}
		api.Log.Debug("Alarm category updated with success, name: " + category.Name)

		c.JSON(http.StatusOK, tools.EmptyRes())
	}
}
