package cost

import (
	"net/http"
	"sync"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/gin-gonic/gin"
)

var getMu sync.Mutex

// Get the current server cost.
// Responses:
//   - 200 If succeeded.
func GetCostHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		getMu.Lock()
		defer getMu.Unlock()

		ctx := c.Request.Context()

		var result models.ServerCostResult
		cacheRes, err := api.Cache.GetServerCostResult(ctx)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to get server cost on cache", logger.ErrField(err))
			return
		}
		if !cacheRes.Exists {
			result, err = calculate(ctx, api)
			if err != nil {
				c.Status(http.StatusInternalServerError)
				api.Log.Error("Fail to calculate server cost", logger.ErrField(err))
				return
			}

			err = api.Cache.SetServerCostResult(ctx, result)
			if err != nil {
				c.Status(http.StatusInternalServerError)
				api.Log.Error("Fail to set server cost on cache", logger.ErrField(err))
				return
			}
		}

		c.JSON(http.StatusOK, result)
	}
}
