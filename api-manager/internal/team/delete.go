package team

import (
	"log"
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/gin-gonic/gin"
)

// Deletes team from databse
// Responses:
//   - 400 If invalid id
//   - 404 If team not founded
//   - 204 If succeeded
func DeleteHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		// get id from param
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// check if team exists
		e, err := api.PgConn.Teams.Exists(c.Request.Context(), id)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			log.Printf("\nfail to query team, err: %s", err)
			return
		}
		if !e {
			c.Status(http.StatusNotFound)
			return
		}

		_, err = api.PgConn.Teams.Delete(c.Request.Context(), id)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			log.Printf("\nfail to delete team, err: %s", err)
			return
		}
		log.Printf("\tteam id %d deleted successfully", id)

		c.Status(http.StatusNoContent)
	}
}
