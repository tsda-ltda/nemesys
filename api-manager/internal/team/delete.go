package team

import (
	"log"
	"net/http"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/gin-gonic/gin"
)

// Deletes team from databse
// Responses:
//   - 404 If team not founded
//   - 204 If succeeded
func DeleteHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		// get id
		id, err := getId(api, c)
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}

		// delete team
		t, err := api.PgConn.Exec(c.Request.Context(), "DELETE FROM teams WHERE id = $1", id)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			log.Printf("fail to delete team, err: %s", err)
			return
		}
		if t.RowsAffected() == 0 {
			c.Status(http.StatusNotFound)
			return
		}

		c.Status(http.StatusNoContent)
	}
}
