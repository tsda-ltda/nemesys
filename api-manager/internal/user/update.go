package user

import (
	"net/http"
	"strconv"

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
//   - 400 If username or email already in use.
//   - 404 If user not founded.
//   - 200 If succeeded.
func UpdateHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// get id param
		id, err := strconv.ParseInt(c.Param("id"), 10, 0)
		if err != nil {
			c.Status((http.StatusBadRequest))
			return
		}

		// bind user
		var user models.User
		err = c.ShouldBind(&user)
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

		// get username and email availability
		ue, ee, err := api.PgConn.Users.UsernameEmailAvailableToUpdate(ctx, int32(id), user.Username, user.Email)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to check if username and email exists", logger.ErrField(err))
			return
		}

		// check if username exists
		if ue {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgUsernameExists))
			return
		}

		// check if email exists
		if ee {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgEmailExists))
			return
		}

		// hash password
		pwHashed, err := auth.Hash(user.Password, api.UserPWBcryptCost)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to hash password", logger.ErrField(err))
			return
		}

		// update user
		e, err := api.PgConn.Users.Update(ctx, models.User{
			Id:       int32(id),
			Role:     user.Role,
			Name:     user.Name,
			Username: user.Username,
			Password: pwHashed,
			Email:    user.Email,
		})
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to update user", logger.ErrField(err))
			return
		}

		// check if user exists
		if e {
			c.Status(http.StatusNotFound)
			return
		}

		api.Log.Debug("user updated with success, username: " + user.Username)
		c.Status(http.StatusOK)
	}
}
