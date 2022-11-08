package user

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
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

		// get id
		id, err := strconv.ParseInt(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		// deltete user
		e, err := api.PgConn.Users.Delete(ctx, int32(id))
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to delete user", logger.ErrField(err))
			return
		}
		if !e {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgUserNotFound))
			return
		}

		api.Log.Debug("user deleted with success, id: " + fmt.Sprint(id))
		c.Status(http.StatusNoContent)
	}
}
