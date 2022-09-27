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

		// get users ids
		sql := `SELECT users_ids FROM teams WHERE id = $1`
		rows, err := api.PgConn.Query(c.Request.Context(), sql, id)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			log.Printf("\nfail to query team id and users ids, err: %s", err)
			return
		}

		// scan results
		var usersIds []int
		for rows.Next() {
			rows.Scan(&usersIds)
		}

		// check if team doesn't exists
		if rows.CommandTag().RowsAffected() == 0 {
			c.Status(http.StatusNotFound)
			return
		}

		// delete team
		_, err = api.PgConn.Teams.Delete(c.Request.Context(), id)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			log.Printf("\nfail to delete team, err: %s", err)
			return
		}

		// remove team id from users
		go updateUsersTeamsIds(usersIds, []int{}, id, api, c.Request.Context())

		c.Status(http.StatusNoContent)
	}
}
