package team

import (
	"log"
	"net/http"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/gin-gonic/gin"
)

// Team struct for MGetHandler json responses
type _MGetTeam struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Ident string `json:"ident"`
	Descr string `json:"descr"`
}

// Get team in database
// Responses:
//   - 404 If team not foud
//   - 200 If succeeded
func GetHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		// get team id
		id, err := getId(api, c)
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}

		// get team
		var team models.Team
		sql := `SELECT users_ids, descr, name, ident FROM teams WHERE id = $1`
		rows, err := api.PgConn.Query(c.Request.Context(), sql, id)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			log.Printf("\nfail to query team, err: %s", err)
			return
		}

		// scan results
		for rows.Next() {
			rows.Scan(
				&team.UsersIds,
				&team.Descr,
				&team.Name,
				&team.Ident,
			)
		}

		// set id
		team.Id = id

		// check if team doesn't exists
		if rows.CommandTag().RowsAffected() == 0 {
			c.Status(http.StatusNotFound)
			return
		}

		c.JSON(http.StatusOK, team)
	}
}

// Get multi teams on databse
// Params:
//   - "limit" Limit of teams returned. Default is 30, max is 30, min is 0.
//   - "offset" Offset for searching. Default is 0, min is 0.
//
// Responses:
//   - 400 If invalid params.
//   - 200 If succeeded.
func MGetHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		// db query params
		limit, err := tools.IntRangeQuery(c, "limit", 30, 30, 1)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		offset, err := tools.IntMinQuery(c, "offset", 0, 0)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// search teams
		sql := `SELECT id, name, descr, ident FROM teams LIMIT $1 OFFSET $2`
		rows, err := api.PgConn.Query(c.Request.Context(), sql, limit, offset)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			log.Printf("\nfail to query teams, err: %s", err)
			return
		}
		defer rows.Close()

		// scan teams
		teams := []_MGetTeam{}
		for rows.Next() {
			var t _MGetTeam
			rows.Scan(&t.Id, &t.Name, &t.Descr, &t.Ident)
			teams = append(teams, t)
		}

		c.JSON(http.StatusOK, teams)
	}
}
