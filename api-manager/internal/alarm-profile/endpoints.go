package profile

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/pg"
	"github.com/gin-gonic/gin"
)

// Creates a alarm endpoint relation.
// Responses:
//   - 400 If invalid params.
//   - 404 If alarm profile not found.
//   - 404 If alarm endpoint not found.
//   - 400 If relation already exists.
//   - 200 If succeeded.
func CreateAlarmEndpointRelation(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		alarmProfileId, err := strconv.ParseInt(c.Param("profileId"), 0, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}

		var alarmEndpoint models.Id32
		err = c.ShouldBind(&alarmEndpoint)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}

		err = api.Validate.Struct(alarmEndpoint)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}

		r, err := api.PG.AlarmEndpointRelationExists(ctx, int32(alarmProfileId), alarmEndpoint.Id)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to check if alarm endpoint relation exists", logger.ErrField(err))
			return
		}

		if r.RelationExists {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgAlarmEndpointRelationExists))
			return
		}
		if !r.AlarmProfileExists {
			c.JSON(http.StatusNotFound, tools.MsgRes(tools.MsgAlarmProfileNotFound))
			return
		}
		if !r.AlarmEndpointExists {
			c.JSON(http.StatusNotFound, tools.MsgRes(tools.MsgAlarmEndpointNotFound))
			return
		}

		err = api.PG.CreateAlarmEndpointRelation(ctx, int32(alarmProfileId), alarmEndpoint.Id)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to create alarm endpoint relation", logger.ErrField(err))
			return
		}

		c.JSON(http.StatusOK, tools.EmptyRes())
	}
}

// Deletes a alarm endpoint relation.
// Responses:
//   - 400 If invalid params.
//   - 404 If not found.
//   - 200 If succeeded.
func DeleteAlarmEndpointRelation(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		alarmProfileId, err := strconv.ParseInt(c.Param("profileId"), 0, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}

		alarmEndpointId, err := strconv.ParseInt(c.Param("endpointId"), 0, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}

		exists, err := api.PG.DeleteAlarmEndpointRelation(ctx, int32(alarmProfileId), int32(alarmEndpointId))
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to create alarm endpoint relation", logger.ErrField(err))
			return
		}

		if !exists {
			c.JSON(http.StatusNotFound, tools.MsgRes(tools.MsgAlarmEndpointNotFound))
			return
		}

		c.JSON(http.StatusOK, tools.EmptyRes())
	}
}

func GetAlarmEndpoints(api *api.API) func(c *gin.Context) {
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

		alarmProfileId, err := strconv.ParseInt(c.Param("profileId"), 0, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}

		endpoints, err := api.PG.GetAlamProfileAlarmEndpoints(ctx, pg.AlarmEndpointQueryFilters{
			AlarmProfileId: int32(alarmProfileId),
			Name:           c.Query("name"),
			URL:            c.Query("url"),
			OrderBy:        c.Query("orderBy"),
			OrderByFn:      c.Query("orderByFn"),
		}, limit, offset)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to create alarm endpoint relation", logger.ErrField(err))
			return
		}

		c.JSON(http.StatusOK, tools.DataRes(endpoints))
	}
}
