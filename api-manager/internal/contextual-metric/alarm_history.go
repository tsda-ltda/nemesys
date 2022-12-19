package ctxmetric

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/influxdb"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/gin-gonic/gin"
)

// Get metric alarm history.
// Responses:
//   - 400 If invalid params.
//   - 200 If succeeded.
func AlarmHistoryHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		req, err := tools.GetMetricRequest(c)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}

		var opts influxdb.QueryAlarmHistoryOptions

		start, err := strconv.ParseInt(c.Query("start"), 0, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}

		rawStop := c.Query("stop")
		if rawStop != "" {
			stop, err := strconv.ParseInt(rawStop, 0, 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
				return
			}
			opts.Stop = stop
		}

		rawLevel := c.Query("level")
		if rawLevel != "" {
			level, err := strconv.ParseInt(rawLevel, 0, 32)
			if err != nil {
				c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
				return
			}
			levelInt32 := int32(level)
			opts.Level = &levelInt32
		}

		opts.ContainerId = req.ContainerId
		opts.MetricId = req.MetricId
		opts.Start = start

		points, err := api.Influx.QueryAlarmHistory(ctx, opts)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to query alarm history", logger.ErrField(err))
			return
		}

		c.JSON(http.StatusOK, tools.DataRes(points))
	}
}
