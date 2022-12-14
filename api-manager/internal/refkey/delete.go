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
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}

		exists, err := api.PG.DeleteMetricRefkey(ctx, rkId)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to delete refkey", logger.ErrField(err))
			return
		}
		if !exists {
			c.JSON(http.StatusNotFound, tools.MsgRes(tools.MsgRefkeyNotFound))
			return
		}

		c.JSON(http.StatusOK, tools.EmptyRes())
	}
}
