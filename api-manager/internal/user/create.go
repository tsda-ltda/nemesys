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

// Create user on database.
// Responses:
//   - 400 If invalid body.
//   - 400 If json fields are invalid.
//   - 400 If username is already in use.
//   - 400 If email is already in use.
//   - 200 If succeeded.
func CreateHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		var user models.User
		err := c.ShouldBind(&user)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidBody))
			return
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

		usernameExists, emailExists, err := api.PG.UsernameAndEmailExists(ctx, user.Username, user.Email, -1)
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

		pwHashed, err := auth.Hash(user.Password, api.UserPWBcryptCost)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to hash password", logger.ErrField(err))
			return
		}

		user.Password = pwHashed
		id, err := api.PG.CreateUser(ctx, user)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to create user", logger.ErrField(err))
			return
		}
		api.Log.Info("User created, id: " + strconv.FormatInt(int64(id), 10))

		c.JSON(http.StatusOK, tools.IdRes(int64(id)))
	}
}
