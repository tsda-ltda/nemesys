package team

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/gin-gonic/gin"
)

// Get a context.
// Responses:
//   - 400 If invalid id.
//   - 404 If id not found
//   - 200 If succeeded.
func GetContextHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// context id
		contextId, err := strconv.ParseInt(c.Param("ctxId"), 10, 32)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// get contexts
		e, context, err := api.PgConn.Contexts.Get(ctx, int32(contextId))
		if err != nil {
			api.Log.Error("fail to get contexts", logger.ErrField(err))
			c.Status(http.StatusInternalServerError)
			return
		}

		// check if not exists
		if !e {
			c.Status(http.StatusNotFound)
			return
		}

		c.JSON(http.StatusOK, context)
	}
}

// Get all team's contexts.
// Params:
//   - "limit" Limit of teams returned. Default is 30, max is 30, min is 0.
//   - "offset" Offset for searching. Default is 0, min is 0.
//
// Responses:
//   - 400 If invalid params.
//   - 200 If succeeded.
func MGetContextHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// team id
		teamId, err := strconv.ParseInt(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		// db query params
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

		// get contexts
		ctxs, err := api.PgConn.Contexts.MGet(ctx, int32(teamId), limit, offset)
		if err != nil {
			api.Log.Error("fail to get contexts", logger.ErrField(err))
			c.Status(http.StatusInternalServerError)
			return
		}

		c.JSON(http.StatusOK, ctxs)
	}
}
