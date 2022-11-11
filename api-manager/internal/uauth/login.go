package uauth

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/auth"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/env"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/gin-gonic/gin"
)

type loginReq struct {
	Username string `json:"username" validate:"required,min=2,max=50"`
	Password string `json:"password" validate:"required,min=5,max=50"`
}

// Login into a user account.
// Responses:
//   - 400 If invalid body.
//   - 400 If invalid body fields.
//   - 404 If username or password is incorrect.
//   - 200 If succeeded.
func LoginHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		var form loginReq

		// bind login form
		err := c.ShouldBind(&form)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidBody))
			return
		}

		// validate
		err = api.Validate.Struct(form)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidJSONFields))
			return
		}

		// get login info
		li, err := api.PgConn.Users.LoginInfo(ctx, form.Username)
		if err != nil {
			api.Log.Error("fail to get login info", logger.ErrField(err))
			c.Status(http.StatusInternalServerError)
			return
		}

		// check if user exists
		if !li.Exists {
			c.JSON(http.StatusUnauthorized, tools.JSONMSG(tools.MsgWrongUsernameOrPW))
			return
		}

		// check password
		if !auth.CheckHash(form.Password, li.Password) {
			c.JSON(http.StatusUnauthorized, tools.JSONMSG(tools.MsgWrongUsernameOrPW))
			return
		}

		// create new session
		token, err := api.Auth.NewSession(ctx, auth.SessionMeta{
			UserId: int32(li.Id),
			Role:   uint8(li.Role),
		})
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to create user session", logger.ErrField(err))
			return
		}

		// save session token in cookie
		ttl, _ := strconv.Atoi(env.UserSessionTTL)
		c.SetCookie(auth.SessionCookieName, token, ttl, "/", env.APIManagerHost, false, true)

		c.Status(http.StatusOK)
	}
}
