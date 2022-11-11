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

		// get session metadata
		meta, err := tools.GetSessionMeta(c)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to get session metadata", logger.ErrField(err))
			return
		}

		// remove session
		err = api.Auth.RemoveSession(ctx, meta.UserId)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgSessionAlreadyRemoved))
			return
		}
		api.Log.Debug(fmt.Sprintf("user '%d' logout with success", meta.UserId))

		c.Status(http.StatusOK)
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

		// get session metadata
		meta, err := tools.GetSessionMeta(c)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to get session metadata", logger.ErrField(err))
			return
		}

		// get user id
		rawId := c.Param("id")
		id, err := strconv.ParseInt(rawId, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		// get user role
		r, err := api.PgConn.Users.GetRole(ctx, int32(id))
		if err != nil {
			api.Log.Error("fail to get user role", logger.ErrField(err))
			c.Status(http.StatusInternalServerError)
			return
		}

		// check if user exists
		if !r.Exists {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgUserNotFound))
			return
		}

		// check if target's role is superior
		if uint8(r.Role) > meta.Role {
			c.Status(http.StatusForbidden)
			return
		}

		// remove session
		err = api.Auth.RemoveSession(ctx, int32(id))
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgSessionAlreadyRemoved))
			return
		}
		api.Log.Debug("user forcibly logout with success, id: " + rawId)
		c.Status(http.StatusOK)
	}
}
