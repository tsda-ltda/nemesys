package customquery

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/pg"
	"github.com/gin-gonic/gin"
)

// Get multi custom queries.
// Params:
//   - "limit" Limit of metrics returned. Default is 30, max is 30, min is 0.
//   - "offset" Offset for searching. Default is 0, min is 0.
//
// Responses:
//   - 400 If invalid params.
//   - 200 If succeeded.
func MGetHandler(api *api.API) func(c *gin.Context) {
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

		cqs, err := api.PG.GetCustomQueries(ctx, pg.CustomQueryQueryFilters{
			Ident:     c.Query("ident"),
			Descr:     c.Query("descr"),
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
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to get multi custom queries on database", logger.ErrField(err))
			return
		}

		c.JSON(http.StatusOK, tools.DataRes(cqs))
	}
}

// GetHandler a custom query.
// Responses:
//   - 400 If invalid params.
//   - 404 If not found.
//   - 200 If succeeded.
func GetHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		id, err := strconv.ParseInt(c.Param("cqId"), 0, 10)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}

		exists, cq, err := api.PG.GetCustomQuery(ctx, int32(id))
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to get custom query on database", logger.ErrField(err))
			return
		}
		if !exists {
			c.JSON(http.StatusNotFound, tools.MsgRes(tools.MsgCustomQueryNotFound))
			return
		}

		c.JSON(http.StatusOK, tools.DataRes(cq))
	}
}
