package middleware

import (
	"net/http"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/roles"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/gin-gonic/gin"
)

func RequestsCounter(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		meta, err := tools.GetSessionMeta(c)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			api.Log.Error("Fail to get session metadata", logger.ErrField(err))
			return
		}
		_, ok := api.Counter.Whitelist.Load(meta.UserId)
		if !ok && meta.Role < roles.Master {
			api.Counter.IncrRequests()
		}
		c.Next()
	}
}

func RealtimeDataRequestsCounter(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		meta, err := tools.GetSessionMeta(c)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			api.Log.Error("Fail to get session metadata", logger.ErrField(err))
			return
		}
		_, ok := api.Counter.Whitelist.Load(meta.UserId)
		if !ok && meta.Role < roles.Master {
			api.Counter.IncrRealtimeDataRequests()
		}
		c.Next()
	}
}

func DataHistoryRequestsCounter(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		meta, err := tools.GetSessionMeta(c)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			api.Log.Error("Fail to get session metadata", logger.ErrField(err))
			return
		}
		_, ok := api.Counter.Whitelist.Load(meta.UserId)
		if !ok && meta.Role < roles.Master {
			api.Counter.IncrDataHistoryRequests()
		}
		c.Next()
	}
}
