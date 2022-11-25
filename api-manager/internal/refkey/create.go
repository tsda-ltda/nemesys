package refkey

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

// Creates a metric reference key.
// Responses:
//   - 400 If invalid body.
//   - 400 If json fields are invalid.
//   - 404 If metric not found.
//   - 400 If refkey already exists.
//   - 200 If succeeded.
func CreateHandler(api *api.API, containerType types.ContainerType) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		metricId, err := strconv.ParseInt(c.Param("metricId"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		var rk models.MetricRefkey
		err = c.ShouldBind(&rk)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidBody))
			return
		}

		err = api.Validate.Struct(rk)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidJSONFields))
			return
		}

		rk.MetricId = metricId
		metricExists, rkExists, err := api.PG.MetricAndRefkeyExists(ctx, metricId, containerType, rk.Refkey, -1)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to check if metric and refkey exists", logger.ErrField(err))
			return
		}
		if !metricExists {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgMetricNotFound))
			return
		}
		if rkExists {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgRefkeyExists))
			return
		}

		_, err = api.PG.CreateMetricRefkey(ctx, rk)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to create refkey", logger.ErrField(err))
			return
		}

		c.Status(http.StatusOK)
	}
}
