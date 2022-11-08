package datapolicy

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/gin-gonic/gin"
)

// Deletes a data policy.
// Responses:
//   - 400 If invalid id.
//   - 404 If data policy not found.
//   - 204 If succeeded.
func DeleteHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// get data policy id
		id, err := strconv.ParseInt(c.Param("id"), 10, 16)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		// delete data policy
		e, err := api.PgConn.DataPolicy.Delete(ctx, int16(id))
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to delete data policy", logger.ErrField(err))
			return
		}

		// check if data policy exists
		if !e {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgDataPolicyNotFound))
			return
		}
		api.Log.Info("data policy deleted, id: " + fmt.Sprint(id))
		c.Status(http.StatusNoContent)
	}
}
