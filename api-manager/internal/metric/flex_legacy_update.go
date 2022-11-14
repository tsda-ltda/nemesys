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

// Updates a Flex Legacy metric.
// Responses:
//   - 400 If invalid body.
//   - 400 If json fields are invalid.
//   - 404 If container or metric not found.
//   - 200 If succeeded.
func UpdateFlexLegacyHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// get container id
		rawContainerId := c.Param("containerId")
		containerId, err := strconv.ParseInt(rawContainerId, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		// get metric id
		rawId := c.Param("metricId")
		id, err := strconv.ParseInt(rawId, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		// bind body
		var metric models.Metric[models.FlexLegacyMetric]
		err = c.ShouldBind(&metric)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidBody))
			return
		}

		// validate body
		err = api.Validate.Struct(metric)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidJSONFields))
			return
		}

		// assign id
		metric.Base.Id = id
		metric.Protocol.Id = id
		metric.Base.ContainerId = int32(containerId)

		// get if container and data policy exists
		r, err := api.PgConn.Metrics.ExistsContainerAndDataPolicy(ctx, metric.Base.ContainerId, metric.Base.DataPolicyId, id)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to check container and data policy existence", logger.ErrField(err))
			return
		}

		// check metric if exists
		if !r.Exists {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgMetricNotFound))
			return
		}

		// check if data policy exists
		if !r.DataPolicyExists {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgDataPolicyNotFound))
			return
		}

		// check if container exists
		if !r.ContainerExists {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgContainerNotFound))
			return
		}

		// create metric
		exists, err := api.PgConn.FlexLegacyMetrics.Update(ctx, metric.Protocol)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to update flex legacy metric", logger.ErrField(err))
			return
		}

		// check if exists
		if !exists {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgMetricNotFound))
			return
		}

		// update metric
		exists, err = api.PgConn.Metrics.Update(ctx, metric.Base)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to update base metric", logger.ErrField(err))
			return
		}

		// check if metric exists
		if !exists {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgMetricNotFound))
			api.Log.Error("flex legacy metic exists but base metric don't")
			return
		}
		api.Log.Debug("metric updated, name" + metric.Base.Name)
		api.Amqph.NotifyMetricUpdated(metric.Base, metric.Protocol, types.CTFlexLegacy)

		c.Status(http.StatusOK)
	}
}
