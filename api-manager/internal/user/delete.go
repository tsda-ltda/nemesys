package user

import (
	"log"
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
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
			log.Printf("fail to delete user, err: %s", err)
			return
		}
		if flag.RowsAffected() == 0 {
			c.Status(http.StatusNotFound)
			return
		}

		log.Printf("user id %d deleted successfully", id)

		c.Status(http.StatusNoContent)
	}
}
