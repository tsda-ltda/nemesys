package team

import (
	"log"
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/gin-gonic/gin"
)

// UserId struct for AddUserHandler json requests.
type _UserId struct {
	UserId int `json:"user-id" validate:"required"`
}

// Add a user to team.
// Responses:
//   - 400 If invalid body.
//   - 400 If invalid user id.
//   - 400 If user is already a member.
//   - 404 If team does not exists.
//   - 200 If succeeded.
func AddUserHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		// get team id
		teamId, err := getId(api, c)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// get user id
		var userId _UserId
		err = c.ShouldBind(&userId)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// validate userid
		err = api.Validate.Struct(userId)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// query realation existence
		var e bool
		sql := `SELECT EXISTS (SELECT 1 FROM users_teams WHERE userId = $1 AND teamId = $2)`
		err = api.PgConn.QueryRow(c.Request.Context(), sql, userId.UserId, teamId).Scan(&e)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			log.Printf("fail to query users_teams, err: %s", err)
			return
		}

		// check if realation already exists
		if e {
			c.Status(http.StatusBadRequest)
			return
		}

		// add user
		sql = `INSERT INTO users_teams (userid, teamid) VALUES ($1,$2)`
		_, err = api.PgConn.Exec(c.Request.Context(), sql, userId.UserId, teamId)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		c.Status(http.StatusOK)
	}
}

// Remove a user from team.
// Responses:
//   - 400 If invalid user id.
//   - 400 If user is already a member.
//   - 404 If relation does not exists.
//   - 204 If succeeded.
func RemoveUserHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		// get team teamId
		teamId, err := getId(api, c)
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}

		// get user id
		userId, err := strconv.Atoi(c.Param("userId"))
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// remove user from team
		sql := `DELETE FROM users_teams WHERE userId = $1 AND teamId = $2`
		t, err := api.PgConn.Exec(c.Request.Context(), sql, userId, teamId)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}

		// check if relation exists
		if t.RowsAffected() == 0 {
			c.Status(http.StatusNotFound)
			return
		}

		c.Status(http.StatusNoContent)
	}
}

// Get user's teams.
// Params:
//   - "limit" Limit of teams returned. Default is 30, max is 30, min is 0.
//   - "offset" Offset for searching. Default is 0, min is 0.
//
// Responses:
//   - 400 If invalid params.
//   - 200 If succeeded.
func UserTeamsHandler(api *api.API) func(c *gin.Context) {
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

		// get user session metadata
		meta, err := tools.GetSessionMeta(c)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			log.Printf("fail to read session metadata, err: %s", err)
			return
		}

		// query teams
		sql := `SELECT id, name, ident, descr FROM users_teams ut LEFT JOIN teams t ON ut.teamId = t.id WHERE ut.userid = $1 LIMIT $2 OFFSET $3;`
		rows, err := api.PgConn.Query(c.Request.Context(), sql, meta.UserId, limit, offset)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			log.Printf("fail to query users teams, err: %s", err)
			return
		}

		// scan rows
		var teams []_SanitizedTeam
		for rows.Next() {
			var t _SanitizedTeam
			rows.Scan(&t.Id, &t.Name, &t.Ident, &t.Descr)
			teams = append(teams, t)
		}
		rows.Close()

		c.JSON(http.StatusOK, teams)
	}
}
