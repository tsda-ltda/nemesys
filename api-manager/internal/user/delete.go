package user

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
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
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// deltete user
		e, err := api.PgConn.Users.Delete(ctx, id)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to delete user", logger.ErrField(err))
			return
		}
		if !e {
			c.Status(http.StatusNotFound)
			return
		}

		api.Log.Debug("user deleted with success, id: " + fmt.Sprint(id))
		c.Status(http.StatusNoContent)
	}
}
