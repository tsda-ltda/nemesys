package team

import (
	"log"
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/gin-gonic/gin"
)

// Team struct for MGetHandler json responses
type _MGetTeam struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Ident string `json:"ident"`
}

// Team struct for GetHandler json responses
type _GetTeam struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	Ident    string `json:"ident"`
	UsersIds []int  `json:"users-ids"`
}

// Get team in database
// Responses:
//   - 400 If invalid id
//   - 404 If team not foud
//   - 200 If succeeded
func GetHandler(api *api.API) func(c *gin.Context) {
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

		// get team
		var team _GetTeam
		sql := `SELECT id, name, ident, users_ids FROM teams WHERE id = $1`
		err = api.PgConn.QueryRow(c.Request.Context(), sql, id).Scan(
			&team.Id,
			&team.Name,
			&team.Ident,
			&team.UsersIds,
		)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			log.Printf("\nfail to query team, err: %s", err)
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
		sql := `SELECT id, name, ident FROM teams LIMIT $1 OFFSET $2`
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
			rows.Scan(&t.Id, &t.Name, &t.Ident)
			teams = append(teams, t)
		}

		c.JSON(http.StatusOK, teams)
	}
}
