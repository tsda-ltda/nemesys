package refkey

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/gin-gonic/gin"
)

// Deletes a metric reference key.
// Responses:
//   - 404 If metric not found.
//   - 200 If succeeded.
func DeleteHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		rkId, err := strconv.ParseInt(c.Param("refkeyId"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		exists, err := api.PG.DeleteMetricRefkey(ctx, rkId)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to delete refkey", logger.ErrField(err))
			return
		}
		if !exists {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgRefkeyNotFound))
			return
		}

		c.Status(http.StatusNoContent)
	}
}
