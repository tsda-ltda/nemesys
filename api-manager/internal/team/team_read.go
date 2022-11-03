package team

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/gin-gonic/gin"
)

// Get team in database
// Responses:
//   - 404 If team not foud
//   - 200 If succeeded
func GetHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// get team id
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// get team
		e, team, err := api.PgConn.Teams.Get(ctx, int32(id))
		if err != nil {
			api.Log.Error("fail to get team", logger.ErrField(err))
			c.Status(http.StatusInternalServerError)
			return
		}

		// check if team  exists
		if !e {
			c.Status(http.StatusNotFound)
			return
		}

		c.JSON(http.StatusOK, team)
	}
}

// Get multi teams on database
// Params:
//   - "limit" Limit of teams returned. Default is 30, max is 30, min is 0.
//   - "offset" Offset for searching. Default is 0, min is 0.
//
// Responses:
//   - 400 If invalid params.
//   - 200 If succeeded.
func MGetHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// db query params
		limit, err := tools.IntRangeQuery(c, "limit", 30, 30, 1)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		offset, err := tools.IntMinQuery(c, "offset", 0, 0)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// get teams
		teams, err := api.PgConn.Teams.MGet(ctx, limit, offset)
		if err != nil {
			api.Log.Error("fail to get teams", logger.ErrField(err))
			c.Status(http.StatusInternalServerError)
			return
		}

		c.JSON(http.StatusOK, teams)
	}
}
