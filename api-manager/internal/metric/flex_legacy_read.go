package metric

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/pg"
	"github.com/gin-gonic/gin"
)

// Get a Flex Legacy metric.
// Responses:
//   - 400 If invalid params.
//   - 404 If not found.
//   - 200 If succeeded.
func GetFlexLegacyHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		id, err := strconv.ParseInt(c.Param("metricId"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}

		exists, metric, err := api.PG.GetFlexLegacyMetric(ctx, id)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to get flex legacy metric", logger.ErrField(err))
			return
		}
		if !exists {
			c.JSON(http.StatusNotFound, tools.MsgRes(tools.MsgMetricNotFound))
			return
		}

		c.JSON(http.StatusOK, tools.DataRes(metric))
	}
}

// Get multi Basic metrics.
// Responses:
//   - 200 If succeeded.
func MGetFlexLegacyHandler(api *api.API) func(c *gin.Context) {
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

		containerId, err := strconv.ParseInt(c.Param("containerId"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}

		var e bool
		var enabled *bool
		rawEnabled := c.Query("enabled")
		if rawEnabled == "1" {
			e = true
			enabled = &e
		} else if rawEnabled == "0" {
			enabled = &e
		}

		port, _ := strconv.ParseInt(c.Query("port"), 0, 16)
		portType, _ := strconv.ParseInt(c.Query("portType"), 0, 16)
		metrics, err := api.PG.GetFlexLegacyMetrics(ctx, pg.FlexLegacyMetricQueryFilters{
			ContainerId: int32(containerId),
			Name:        c.Query("name"),
			Descr:       c.Query("descr"),
			Enabled:     enabled,
			Port:        int16(port),
			PortType:    int16(portType),
			OrderBy:     c.Query("order-by"),
			OrderByFn:   c.Query("order-by-fn"),
		}, limit, offset)
		if err != nil {
			if err == pg.ErrInvalidOrderByColumn || err == pg.ErrInvalidFilterValue || err == pg.ErrInvalidOrderByFn {
				c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
				return
			}
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to get flex legacy metrics", logger.ErrField(err))
			return
		}

		c.JSON(http.StatusOK, tools.DataRes(metrics))
	}
}
