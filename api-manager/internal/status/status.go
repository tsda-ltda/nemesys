package status

import (
	"net/http"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/gin-gonic/gin"
)

// Get services status.
// Responses:
//   - 200 If succeeded.
func GetHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, tools.DataRes(api.GetServicesStatus()))
	}
}
