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

// Creates an trap category relation.
// Responses:
//   - 400 If invalid params.
//   - 404 If alarm category not found.
//   - 200 If succeeded.
func CreateTrapRelationHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		var relation models.TrapCategoryRelation
		err := c.ShouldBind(&relation)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidBody))
			return
		}

		err = api.Validate.Struct(relation)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidJSONFields))
			return
		}

		exists, err := api.PG.AlarmCategoryExists(ctx, relation.AlarmCategoryId)
		if err != nil {
			api.Log.Error("Fail to check if alarm category exists", logger.ErrField(err))
			return
		}
		if !exists {
			c.JSON(http.StatusNotFound, tools.MsgRes(tools.MsgAlarmCategoryNotFound))
			return
		}

		exists, err = api.PG.TrapCategoryRelationExists(ctx, relation.TrapCategoryId)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to check if trap relation exists", logger.ErrField(err))
			return
		}
		if exists {
			c.JSON(http.StatusNotFound, tools.MsgRes(tools.MsgTrapRelationExists))
			return
		}

		err = api.PG.CreateTrapCategoryRelation(ctx, relation)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to create trap category relation", logger.ErrField(err))
			return
		}
		api.Log.Info("Trap category relation created, category id: " + strconv.Itoa(int(relation.AlarmCategoryId)))

		c.JSON(http.StatusOK, tools.EmptyRes())
	}
}

// Deletes a trap category relation.
// Responses:
//   - 400 If invalid params.
//   - 404 If relation not found.
//   - 200 If succeeded.
func DeleteTrapRelationHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		rawTrapId := c.Param("trapId")
		trapId, err := strconv.ParseInt(rawTrapId, 0, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}

		exists, err := api.PG.DeleteTrapCategoryRelation(ctx, int16(trapId))
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to delete trap relation", logger.ErrField(err))
			return
		}
		if !exists {
			c.JSON(http.StatusNotFound, tools.MsgRes(tools.MsgTrapRelationNotFound))
			return
		}
		api.Log.Info("Trap relation removed, trap id: " + rawTrapId)

		c.JSON(http.StatusOK, tools.EmptyRes())
	}
}

// Get all trap categories relations.
// Params:
//   - "limit" Limit of containers returned. Default is 30, max is 30, min is 0.
//   - "offset" Offset for searching. Default is 0, min is 0.
//
// Responses:
//   - 400 If invalid params.
//   - 200 If succeeded.
func GetTrapRelationsHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		limit, err := tools.IntRangeQuery(c, "limit", 30, 30, 1)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}
		offset, err := tools.IntMinQuery(c, "offset", 0, 0)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}

		emails, err := api.PG.GetTrapCategoriesRelations(ctx, limit, offset)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to get emails from alarm profile", logger.ErrField(err))
			return
		}

		c.JSON(http.StatusOK, tools.DataRes(emails))
	}
}
