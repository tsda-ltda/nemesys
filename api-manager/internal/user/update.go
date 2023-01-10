package user

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/auth"
	"github.com/fernandotsda/nemesys/api-manager/internal/roles"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/gin-gonic/gin"
)

// Updates user on database.
// Responses:
//   - 400 If invalid body.
//   - 400 If json fields are invalid.
//   - 400 If username or email already in use.
//   - 404 If user not found.
//   - 200 If succeeded.
func UpdateHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		rawId := c.Param("userId")
		id, err := strconv.ParseInt(rawId, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}

		var user models.User
		err = c.ShouldBind(&user)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidBody))
			return
		}

		keepPw := false
		if user.Password == "" {
			keepPw = true
			user.Password = "placeholder-to-pass-validation"
		}

		err = api.Validate.Struct(user)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidJSONFields))
			return
		}

		if !roles.ValidateRole(user.Role) {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidRole))
			return
		}

		meta, err := tools.GetSessionMeta(c)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to get user metadata", logger.ErrField(err))
			return
		}

		if meta.UserId == int32(id) {
			user.Role = meta.Role
		} else if meta.Role != roles.Master && meta.Role <= user.Role {
			c.Status(http.StatusForbidden)
			return
		}

		usernameExists, emailExists, err := api.PG.UsernameAndEmailExists(ctx, user.Username, user.Email, int32(id))
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to check if username and email exists", logger.ErrField(err))
			return
		}

		if usernameExists {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgUsernameExists))
			return
		}

		if emailExists {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgEmailExists))
			return
		}

		user.Id = int32(id)
		var exists bool
		if !keepPw {
			pwHashed, err := auth.Hash(user.Password, api.UserPWBcryptCost)
			if err != nil {
				c.Status(http.StatusInternalServerError)
				api.Log.Error("Fail to hash password", logger.ErrField(err))
				return
			}
			user.Password = pwHashed
			exists, err = api.PG.UpdateUser(ctx, user)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				c.Status(http.StatusInternalServerError)
				api.Log.Error("Fail to update user", logger.ErrField(err))
				return
			}
		} else {
			exists, err = api.PG.UpdateUserKeepPW(ctx, user)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				c.Status(http.StatusInternalServerError)
				api.Log.Error("Fail to update user", logger.ErrField(err))
				return
			}
		}
		if !exists {
			c.JSON(http.StatusNotFound, tools.MsgRes(tools.MsgUserNotFound))
			return
		}

		api.Log.Info("User updated, id: " + rawId)
		c.JSON(http.StatusOK, tools.EmptyRes())
	}
}
