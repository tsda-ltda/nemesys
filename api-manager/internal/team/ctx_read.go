package team

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/pg"
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

		contextId, err := strconv.ParseInt(c.Param("ctxId"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}

		exists, context, err := api.PG.GetContext(ctx, int32(contextId))
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			api.Log.Error("Fail to get context", logger.ErrField(err))
			c.Status(http.StatusInternalServerError)
			return
		}
		if !exists {
			c.JSON(http.StatusNotFound, tools.MsgRes(tools.MsgContextNotFound))
			return
		}

		c.JSON(http.StatusOK, tools.DataRes(context))
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

		teamId, err := strconv.ParseInt(c.Param("teamId"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}
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

		ctxs, err := api.PG.GetContexts(ctx, pg.ContextQueryFilters{
			TeamId:    int32(teamId),
			Name:      c.Query("name"),
			Descr:     c.Query("descr"),
			Ident:     c.Query("ident"),
			OrderBy:   c.Query("order-by"),
			OrderByFn: c.Query("order-by-fn"),
			Limit:     limit,
			Offset:    offset,
		})
		if err != nil {
			if err == pg.ErrInvalidOrderByColumn || err == pg.ErrInvalidFilterValue || err == pg.ErrInvalidOrderByFn {
				c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
				return
			}
			if ctx.Err() != nil {
				return
			}
			api.Log.Error("Fail to get contexts", logger.ErrField(err))
			c.Status(http.StatusInternalServerError)
			return
		}

		c.JSON(http.StatusOK, tools.DataRes(ctxs))
	}
}
