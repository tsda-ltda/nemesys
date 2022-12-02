package profile

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/gin-gonic/gin"
)

// Get a alarm profile.
// Responses:
//   - 400 If invalid params.
//   - 404 If not found.
//   - 200 If succeeded.
func GetHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		rawId := c.Param("alarmProfileId")
		id, err := strconv.ParseInt(rawId, 0, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		exists, profile, err := api.PG.GetAlarmProfile(ctx, int32(id))
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to get alarm profile", logger.ErrField(err))
			return
		}
		if !exists {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgAlarmProfileNotFound))
			return
		}

		c.JSON(http.StatusOK, profile)
	}
}

// Get multi alarm profiles.
// Responses:
//   - 400 If invalid params.
//   - 200 If succeeded.
func MGetHandler(api *api.API) func(c *gin.Context) {
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

		profiles, err := api.PG.GetAlarmProfiles(ctx, limit, offset)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to get multi alarm profile", logger.ErrField(err))
			return
		}

		c.JSON(http.StatusOK, profiles)
	}
}
