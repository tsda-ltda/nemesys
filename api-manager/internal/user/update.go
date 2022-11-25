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

		id, err := strconv.ParseInt(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		var user models.User
		err = c.ShouldBind(&user)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidBody))
			return
		}

		err = api.Validate.Struct(user)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidJSONFields))
			return
		}

		r, err := api.PG.UsernameAndEmailExists(ctx, user.Username, user.Email, int32(id))
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to check if username and email exists", logger.ErrField(err))
			return
		}

		if r.UsernameExists {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgUsernameExists))
			return
		}

		if r.EmailExists {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgEmailExists))
			return
		}

		pwHashed, err := auth.Hash(user.Password, api.UserPWBcryptCost)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to hash password", logger.ErrField(err))
			return
		}

		exists, err := api.PG.UpdateUser(ctx, models.User{
			Id:       int32(id),
			Role:     user.Role,
			Name:     user.Name,
			Username: user.Username,
			Password: pwHashed,
			Email:    user.Email,
		})
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to update user", logger.ErrField(err))
			return
		}
		if exists {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgUserNotFound))
			return
		}

		api.Log.Debug("user updated with success, username: " + user.Username)
		c.Status(http.StatusOK)
	}
}
