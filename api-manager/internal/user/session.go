package user

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/gin-gonic/gin"
)

func SessionInfoHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		meta, err := tools.GetSessionMeta(c)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to get session metadata", logger.ErrField(err))
			return
		}

		exists, user, err := api.PG.GetUserWithoutPW(ctx, meta.UserId)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to get user in database", logger.ErrField(err))
			return
		}
		if !exists {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgUserNotFound))
			api.Log.Warn("User does not exists, but have an alive session, user id: " + strconv.FormatInt(int64(meta.UserId), 10))
			return
		}

		c.JSON(http.StatusOK, user)
	}
}
