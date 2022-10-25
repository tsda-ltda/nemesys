package user

import (
	"net/http"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/auth"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/gin-gonic/gin"
)

// Creates a new user on database.
// Responses:
//   - 400 If invalid body.
//   - 400 If json fields are invalid.
//   - 400 If username is already in use.
//   - 400 If email is already in use.
//   - 200 If succeeded.
func CreateHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// bind user
		var user models.User
		err := c.ShouldBind(&user)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// validate user
		err = api.Validate.Struct(user)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// check if username and email exists
		ue, ee, err := api.PgConn.Users.ExistsUsernameEmail(ctx, user.Username, user.Email)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to check if username and email exists", logger.ErrField(err))
			return
		}

		// check if username is in use
		if ue {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgUsernameExists))
			return
		}

		// check if email is in use
		if ee {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgUsernameExists))
			return
		}

		// hash password
		pwHashed, err := auth.Hash(user.Password, api.UserPWBcryptCost)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to hash password", logger.ErrField(err))
			return
		}

		// save user in database
		err = api.PgConn.Users.Create(ctx, models.User{
			Role:     user.Role,
			Name:     user.Name,
			Username: user.Username,
			Password: pwHashed,
			Email:    user.Email,
		})
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to create user", logger.ErrField(err))
			return
		}
		api.Log.Debug("new user created, username: " + user.Username)

		c.Status(http.StatusOK)
	}
}
