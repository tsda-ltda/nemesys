package user

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/gin-gonic/gin"
)

// Get user's teams.
// Params:
//   - "limit" Limit of teams returned. Default is 30, max is 30, min is 0.
//   - "offset" Offset for searching. Default is 0, min is 0.
//
// Responses:
//   - 400 If invalid params.
//   - 200 If succeeded.
func TeamsHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		id, err := strconv.ParseInt(c.Param("userId"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

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

		teams, err := api.PG.GetUserTeams(ctx, int32(id), limit, offset)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			api.Log.Error("Fail to get user's teams", logger.ErrField(err))
			c.Status(http.StatusInternalServerError)
			return
		}

		c.JSON(http.StatusOK, teams)
	}
}
