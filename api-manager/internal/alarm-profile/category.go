package profile

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/gin-gonic/gin"
)

// Add an alarm category to alarm profile.
// Responses:
//   - 400 If invalid params.
//   - 404 If alarm profile not found.
//   - 404 If alarm category not found.
//   - 400 If relation alredy exists.
//   - 200 If succeeded.
func AddCategoryHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		profileId, err := strconv.ParseInt(c.Param("profileId"), 0, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		var id32 models.Id32
		err = c.ShouldBind(&id32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidJSONFields))
			return
		}

		r, err := api.PG.CategoryAndAlarmProfileRelationExists(ctx, int32(profileId), id32.Id)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to check if alarm profile, category and relation exists", logger.ErrField(err))
			return
		}
		if !r.Exists {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgAlarmProfileNotFound))
			return
		}
		if !r.CategoryExists {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgAlarmCategoryNotFound))
			return
		}
		if r.RelationExists {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgAlarmProfileAndCategoryRelExists))
			return
		}

		err = api.PG.AddCategoryToAlarmProfile(ctx, int32(profileId), id32.Id)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to add category to alarm profile", logger.ErrField(err))
			return
		}
		api.Log.Debug("New alarm category added to alarm profile, profile id: " + strconv.FormatInt(profileId, 10))

		c.Status(http.StatusOK)
	}
}

// Remove an alarm category from alarm profile.
// Responses:
//   - 400 If invalid params.
//   - 200 If succeeded.
func RemoveCategoryHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		profileId, err := strconv.ParseInt(c.Param("profileId"), 0, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		categoryId, err := strconv.ParseInt(c.Param("categoryId"), 0, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		exists, err := api.PG.RemoveCategoryFromAlarmProfile(ctx, int32(profileId), int32(categoryId))
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to remove category from alarm profile", logger.ErrField(err))
			return
		}
		if !exists {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgAlarmProfileAndCategoryRelNotFound))
			return
		}
		api.Log.Debug("Alarm category and alarm profile relation deleted, profile id: " + strconv.FormatInt(profileId, 10))

		c.Status(http.StatusNoContent)
	}
}

// Get alarm categories.
// Responses:
//   - 400 If invalid params.
//   - 200 If succeeded.
func GetCategoriesHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		limit, err := tools.IntRangeQuery(c, "limit", 30, 30, 1)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}
		offset, err := tools.IntMinQuery(c, "offset", 0, 0)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		profileId, err := strconv.ParseInt(c.Param("profileId"), 0, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		categories, err := api.PG.GetAlarmProfileCategories(ctx, int32(profileId), limit, offset)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to multi alarm profile categories", logger.ErrField(err))
			return
		}

		c.JSON(http.StatusOK, categories)
	}
}
