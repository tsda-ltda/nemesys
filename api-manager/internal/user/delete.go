package user

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/roles"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/gin-gonic/gin"
)

// Deletes user from databse
// Responses:
//   - 400 If invalid id
//   - 404 If user not founded
//   - 201 If succeeded
func DeleteHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		id, err := strconv.ParseInt(c.Param("userId"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}

		meta, err := tools.GetSessionMeta(c)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to get user metadata", logger.ErrField(err))
			return
		}

		exists, role, err := api.PG.GetUserRole(ctx, int32(id))
		if err != nil {
			api.Log.Error("Fail to get user role", logger.ErrField(err))
			c.Status(http.StatusInternalServerError)
			return
		}
		if !exists {
			c.JSON(http.StatusNotFound, tools.MsgRes(tools.MsgUserNotFound))
			return
		}

		if (meta.Role != roles.Master && roles.Role(role) >= meta.Role) || meta.UserId == int32(id) {
			c.Status(http.StatusForbidden)
			return
		}

		_, err = api.PG.DeleteUser(ctx, int32(id))
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to delete user", logger.ErrField(err))
			return
		}

		api.Log.Info("User deleted, id: " + fmt.Sprint(id))
		c.JSON(http.StatusOK, tools.EmptyRes())
	}
}
