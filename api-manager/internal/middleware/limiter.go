package middleware

import (
	"net/http"
	"time"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/gin-gonic/gin"
)

// Limiter limits the number of request of a client.
// Responses:
//   - 429 If user is suspended.
func Limiter(api *api.API, duration time.Duration) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		ip := c.ClientIP()
		route := c.FullPath()

		suspended, err := api.Cache.GetUserLimited(ctx, ip, route)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.AbortWithStatus(http.StatusInternalServerError)
			api.Log.Error("Fail to get user limited on cache", logger.ErrField(err))
			return
		}
		if suspended {
			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		err = api.Cache.SetUserLimited(ctx, ip, route, duration)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.AbortWithStatus(http.StatusInternalServerError)
			api.Log.Error("Fail to set user limited on cache", logger.ErrField(err))
			return
		}
		c.Next()
	}
}
