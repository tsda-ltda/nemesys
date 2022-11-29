package metric

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/types"
	"github.com/gin-gonic/gin"
)

// Updates a basic metric.
// Responses:
//   - 400 If invalid body.
//   - 400 If json fields are invalid.
//   - 404 If container or metric not found.
//   - 200 If succeeded.
func UpdateBasicHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		rawContainerId := c.Param("containerId")
		containerId, err := strconv.ParseInt(rawContainerId, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		rawId := c.Param("metricId")
		id, err := strconv.ParseInt(rawId, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		var metric models.Metric[struct{}]
		err = c.ShouldBind(&metric)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidBody))
			return
		}

		err = api.Validate.Struct(metric)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidJSONFields))
			return
		}

		metric.Base.Id = id
		metric.Base.ContainerId = int32(containerId)
		metric.Base.ContainerType = types.CTBasic

		r, err := api.PG.MetricContainerAndDataPolicyExists(ctx, metric.Base)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to check container and data policy existence", logger.ErrField(err))
			return
		}
		if !r.Exists {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgMetricNotFound))
			return
		}
		if !r.DataPolicyExists {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgDataPolicyNotFound))
			return
		}
		if !r.ContainerExists {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgContainerNotFound))
			return
		}

		exists, err := api.PG.UpdateBasicMetric(ctx, metric)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to update flex legacy metric", logger.ErrField(err))
			return
		}
		if !exists {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgMetricNotFound))
			return
		}
		api.Log.Debug("metric updated, name" + metric.Base.Name)
		api.Amqph.NotifyMetricUpdated(metric.Base, metric.Protocol, types.CTFlexLegacy)

		c.Status(http.StatusOK)
	}
}