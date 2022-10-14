package datapolicy

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
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
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// delete data policy
		e, err := api.PgConn.DataPolicy.Delete(ctx, id)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to delete data policy", logger.ErrField(err))
			return
		}

		// check if data policy exists
		if !e {
			c.Status(http.StatusNotFound)
			return
		}
		api.Log.Info("data policy deleted, id: " + fmt.Sprint(id))
		c.Status(http.StatusNoContent)
	}
}
