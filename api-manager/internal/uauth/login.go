package uauth

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/auth"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/env"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/gin-gonic/gin"
)

// Login into a user account.
// Responses:
//   - 400 If invalid body.
//   - 400 If invalid body fields.
//   - 404 If username or password is incorrect.
//   - 200 If succeeded.
func LoginHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		var form models.Login

		err := c.ShouldBind(&form)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidBody))
			return
		}

		err = api.Validate.Struct(form)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidJSONFields))
			return
		}

		r, err := api.PG.GetLoginInfo(ctx, form.Username)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			api.Log.Error("Fail to get login info", logger.ErrField(err))
			c.Status(http.StatusInternalServerError)
			return
		}
		if !r.Exists {
			c.JSON(http.StatusUnauthorized, tools.MsgRes(tools.MsgWrongUsernameOrPW))
			return
		}

		if !auth.CheckHash(form.Password, r.Password) {
			c.JSON(http.StatusUnauthorized, tools.MsgRes(tools.MsgWrongUsernameOrPW))
			return
		}

		token, err := api.Auth.NewSession(ctx, auth.SessionMeta{
			UserId: int32(r.Id),
			Role:   uint8(r.Role),
		})
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to create user session", logger.ErrField(err))
			return
		}
		ttl, _ := strconv.Atoi(env.UserSessionTTL)

		c.SetCookie(auth.SessionCookieName, token, ttl, "/", env.APIManagerCookieDomain, false, true)

		c.JSON(http.StatusOK, tools.EmptyRes())
	}
}
