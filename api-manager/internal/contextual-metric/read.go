package ctxmetric

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/gin-gonic/gin"
)

// Get multi contextual metrics on database
// Params:
//   - "limit" Limit of metrics returned. Default is 30, max is 30, min is 0.
//   - "offset" Offset for searching. Default is 0, min is 0.
//
// Responses:
//   - 400 If invalid params.
//   - 200 If succeeded.
func Get(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// contextual metric id
		id, err := strconv.Atoi(c.Param("metricId"))
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// get metric
		e, metric, err := api.PgConn.ContextualMetrics.Get(ctx, int64(id))
		if err != nil {
			api.Log.Error("fail to get metrics", logger.ErrField(err))
			c.Status(http.StatusInternalServerError)
			return
		}

		// check if exists
		if !e {
			c.Status(http.StatusNotFound)
			return
		}

		c.JSON(http.StatusOK, metric)
	}
}

// Get multi contextual metrics on database
// Params:
//   - "limit" Limit of metrics returned. Default is 30, max is 30, min is 0.
//   - "offset" Offset for searching. Default is 0, min is 0.
//
// Responses:
//   - 400 If invalid params.
//   - 200 If succeeded.
func MGet(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// context id
		ctxId, err := strconv.Atoi(c.Param("ctxId"))
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

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

		// check if context exists
		e, err := api.PgConn.Contexts.Exists(ctx, int32(ctxId))
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to check if context exists", logger.ErrField(err))
			return
		}
		if !e {
			c.Status(http.StatusNotFound)
			return
		}

		// get metrics
		metrics, err := api.PgConn.ContextualMetrics.MGet(ctx, int32(ctxId), limit, offset)
		if err != nil {
			api.Log.Error("fail to get metrics", logger.ErrField(err))
			c.Status(http.StatusInternalServerError)
			return
		}

		c.JSON(http.StatusOK, metrics)
	}
}
