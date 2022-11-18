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

// Creates a Flex Legacy metric .
// Responses:
//   - 400 If invalid body.
//   - 400 If json fields are invalid.
//   - 404 If container not found.
//   - 200 If succeeded.
func CreateFlexLegacyHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// container id
		containerId, err := strconv.ParseInt(c.Param("containerId"), 10, 32)
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

		// assign containerId and container type
		metric.Base.ContainerId = int32(containerId)
		metric.Base.ContainerType = types.CTFlexLegacy

		// get if container and data policy exists
		r, err := api.PgConn.Metrics.ExistsContainerAndDataPolicy(ctx, metric.Base)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to check container and data policy existence", logger.ErrField(err))
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

		id, err := api.PgConn.Metrics.Create(ctx, metric.Base)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to create base metric", logger.ErrField(err))
			return
		}

		// assign id
		metric.Protocol.Id = id

		err = api.PgConn.FlexLegacyMetrics.Create(ctx, metric.Protocol)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to create flex legacy metric", logger.ErrField(err))
			return
		}
		api.Log.Debug("flex legacy metric created, name: " + metric.Base.Name)
		api.Amqph.NotifyMetricCreated(metric.Base, metric.Protocol, types.CTFlexLegacy)

		c.Status(http.StatusOK)
	}
}
