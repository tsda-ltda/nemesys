package metric

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	t "github.com/fernandotsda/nemesys/shared/amqph/tools"
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
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}

		rawId := c.Param("metricId")
		id, err := strconv.ParseInt(rawId, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}

		var metric models.Metric[struct{}]
		err = c.ShouldBind(&metric)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidBody))
			return
		}

		err = api.Validate.Struct(metric)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidJSONFields))
			return
		}

		if !types.ValidateMetricType(metric.Base.Type) {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidMetricType))
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
			api.Log.Error("Fail to check container and data policy existence", logger.ErrField(err))
			return
		}
		if !r.Exists {
			c.JSON(http.StatusNotFound, tools.MsgRes(tools.MsgMetricNotFound))
			return
		}
		if !r.DataPolicyExists {
			c.JSON(http.StatusNotFound, tools.MsgRes(tools.MsgDataPolicyNotFound))
			return
		}
		if !r.ContainerExists {
			c.JSON(http.StatusNotFound, tools.MsgRes(tools.MsgContainerNotFound))
			return
		}

		exists, err := api.PG.UpdateBasicMetric(ctx, metric)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to update flex legacy metric", logger.ErrField(err))
			return
		}
		if !exists {
			c.JSON(http.StatusNotFound, tools.MsgRes(tools.MsgMetricNotFound))
			return
		}
		api.Log.Debug("Metric updated, name" + metric.Base.Name)
		t.NotifyMetricUpdated(api.Amqph, metric.Base, metric.Protocol)

		c.JSON(http.StatusOK, tools.EmptyRes())
	}
}
