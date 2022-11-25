package refkey

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/gin-gonic/gin"
)

// Get all references key of a metric.
// Responses:
//   - 404 If metric not found.
//   - 200 If succeeded.
func MGetHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		metricId, err := strconv.ParseInt(c.Param("metricId"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		rks, err := api.PG.GetMetricRefkeys(ctx, metricId)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to delete refkey", logger.ErrField(err))
			return
		}

		c.JSON(http.StatusNoContent, rks)
	}
}

// Get a metric reference key.
// Responses:
//   - 404 If metric not found.
//   - 200 If succeeded.
func GetHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		exists, rk, err := api.PG.GetRefkey(ctx, c.Param("refkey"))
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to delete refkey", logger.ErrField(err))
			return
		}
		if !exists {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgRefkeyNotFound))
			return
		}

		c.JSON(http.StatusNoContent, rk)
	}
}
