package uauth

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/gin-gonic/gin"
)

// Logout of a user account.
// Responses:
//   - 400 If no session was running.
//   - 200 If succeeded.
//
// Keys dependencies:
//   - "sess_meta" Session metadata.
func Logout(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		meta, err := tools.GetSessionMeta(c)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to get session metadata", logger.ErrField(err))
			return
		}
		err = api.Auth.RemoveSession(ctx, meta.UserId)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgSessionAlreadyRemoved))
			return
		}
		api.Log.Debug(fmt.Sprintf("user '%d' logout with success", meta.UserId))

		c.JSON(http.StatusOK, tools.EmptyRes())
	}
}

// Force a user logout.
// Responses:
//   - 400 If invalid id.
//   - 400 If no session was running.
//   - 200 If succeeded.
func ForceLogout(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		meta, err := tools.GetSessionMeta(c)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to get session metadata", logger.ErrField(err))
			return
		}

		rawId := c.Param("userId")
		id, err := strconv.ParseInt(rawId, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}

		exists, role, err := api.PG.GetUserRole(ctx, int32(id))
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			api.Log.Error("Fail to get user role", logger.ErrField(err))
			c.Status(http.StatusInternalServerError)
			return
		}

		if !exists {
			c.JSON(http.StatusNotFound, tools.MsgRes(tools.MsgUserNotFound))
			return
		}
		if uint8(role) > meta.Role {
			c.Status(http.StatusForbidden)
			return
		}

		err = api.Auth.RemoveSession(ctx, int32(id))
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgSessionAlreadyRemoved))
			return
		}
		api.Log.Debug("User forcibly logout with success, id: " + rawId)
		c.JSON(http.StatusOK, tools.EmptyRes())
	}
}
