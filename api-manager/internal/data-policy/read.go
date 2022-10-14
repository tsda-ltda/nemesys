package datapolicy

import (
	"net/http"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/gin-gonic/gin"
)

// Get all data policies.
// Responses:
//   - 200 If succeeded.
func MGetHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// get data policies
		dps, err := api.PgConn.DataPolicy.MGet(ctx)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to read data policies", logger.ErrField(err))
			return
		}
		c.JSON(http.StatusOK, dps)
	}
}
