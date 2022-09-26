package team

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/gin-gonic/gin"
)

// TeamUsers struct for UsersHandler json requests.
type _TeamUsers struct {
	UsersIds []int `json:"users-ids" validate:"required,max=1000"`
}

// Updates a team - users relation on databse.
// Responses:
//   - 400 If invalid id.
//   - 400 If invalid body.
//   - 400 If json is invalid or have duplicated elements.
//   - 404 If team does not exists.
//   - 200 If succeeded.
func UsersHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		// get team id
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		var users _TeamUsers
		err = c.ShouldBind(&users)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// validate users
		err = api.Validate.Struct(users)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// check if has duplicated users
		for i, id := range users.UsersIds {
			for ii, idd := range users.UsersIds {
				if i != ii && id == idd {
					c.Status(http.StatusBadRequest)
					return
				}
			}
		}

		// get prev users
		var prevUsers []int
		sql := `SELECT users_ids FROM teams WHERE id = $1`
		err = api.PgConn.QueryRow(c.Request.Context(), sql, id).Scan(&prevUsers)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			log.Printf("\nfail to query users_ids from teams, err: %s", err)
			return
		}

		// check if team exists
		if prevUsers == nil {
			c.Status(http.StatusNotFound)
			return
		}

		// find added users
		added := []int{}
		for _, new := range users.UsersIds {
			e := false
			for _, old := range prevUsers {
				if new == old {
					e = true
					break
				}
			}
			if !e {
				added = append(added, new)
			}
		}

		// find removed users
		removed := []int{}
		for _, old := range prevUsers {
			e := false
			for _, new := range users.UsersIds {
				if new == old {
					e = true
					break
				}
			}
			if !e {
				removed = append(removed, old)
			}
		}

		// check if no member was added or removed
		if len(added) == 0 && len(removed) == 0 {
			c.Status(http.StatusOK)
			return
		}

		// set team's users_ids
		sql = `UPDATE teams SET users_ids = $1 WHERE id = $2`
		_, err = api.PgConn.Exec(c.Request.Context(), sql, users.UsersIds, id)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			log.Printf("\nfail to update users_ids in team, err: %s", err)
			return
		}

		// update users's teams ids asynchronously
		go updateUsersTeamsIds(removed, added, id, api, c.Request.Context())
		c.Status(http.StatusOK)
	}
}

// Update users teams relation when request context is done
func updateUsersTeamsIds(removed []int, added []int, team int, api *api.API, reqctx context.Context) {
	// wait api send response
	<-reqctx.Done()
	ctx := context.Background()

	// remove team id for old users
	sql := `UPDATE users SET teams_ids = array_remove(teams_ids, $2) WHERE id = $1`
	for _, id := range removed {
		_, err := api.PgConn.Exec(ctx, sql, id, team)
		if err != nil {
			log.Printf("\nfail to remove team id from user, userid: %d, err: %s", id, err)
		}
	}

	// save team id for new users
	sql = `UPDATE users SET teams_ids = array_append(teams_ids, $2) WHERE id = $1`
	for _, id := range added {
		_, err := api.PgConn.Exec(ctx, sql, id, team)
		if err != nil {
			log.Printf("\nfail to save team id in user, userid: %d, err: %s", id, err)
		}
	}
}
