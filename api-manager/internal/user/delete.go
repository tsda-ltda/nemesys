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
		// get id from param
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		flag, err := api.PgConn.Exec(c.Request.Context(), "DELETE FROM users WHERE id = $1", id)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to delete user", logger.ErrField(err))
			return
		}
		if flag.RowsAffected() == 0 {
			c.Status(http.StatusNotFound)
			return
		}

		api.Log.Debug(fmt.Sprintf("user id %d deleted with success", id))
		c.Status(http.StatusNoContent)
	}
}
