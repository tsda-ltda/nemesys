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

// Creates a basic metric .
// Responses:
//   - 400 If invalid body.
//   - 400 If json fields are invalid.
//   - 404 If container not found.
//   - 200 If succeeded.
func CreateBasicHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		containerId, err := strconv.ParseInt(c.Param("containerId"), 10, 32)
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
		if !r.DataPolicyExists {
			c.JSON(http.StatusNotFound, tools.MsgRes(tools.MsgDataPolicyNotFound))
			return
		}
		if !r.ContainerExists {
			c.JSON(http.StatusNotFound, tools.MsgRes(tools.MsgContainerNotFound))
			return
		}

		id, err := api.PG.CreateBasicMetric(ctx, metric)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to create basic metric", logger.ErrField(err))
			return
		}
		api.Log.Debug("Basic metric created, name: " + metric.Base.Name)
		t.NotifyMetricCreated(api.Amqph, metric.Base, metric.Protocol)

		c.JSON(http.StatusOK, tools.IdRes(id))
	}
}
