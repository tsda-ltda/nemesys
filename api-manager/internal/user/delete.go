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
//   - 400 If user doens't exists
//   - 201 If succeeded
func DeleteHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		// get id from param
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// check if user e
		e, err := api.PgConn.Users.Exists(c.Request.Context(), id)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			log.Printf("\nfail to query user, err: %s", err)
			return
		}
		if !e {
			c.Status(http.StatusNotFound)
			return
		}

		_, err = api.PgConn.Users.Delete(c.Request.Context(), id)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			log.Printf("\nfail to delete user, err: %s", err)
			return
		}
		log.Printf("\nuser id %d deleted successfully", id)

		c.Status(http.StatusNoContent)
	}
}
