package user

import (
	"log"
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/gin-gonic/gin"
)

// User struct for MGetHandler json responses
type _MGetUser struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
}

// User struct for GetHandler json responses
type _GetUser struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	TeamsIds []int  `json:"teams-ids"`
}

// Get user in database
// Responses:
//   - 400 If invalid id
//   - 404 If user not foud
//   - 200 If succeeded, containing the users
func GetHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		// get id from param
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// check if user exists
		e, err := api.PgConn.Users.Exists(c.Request.Context(), id)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			log.Printf("\nfail to query user existence, err: %s", err)
			return
		}
		if !e {
			c.Status(http.StatusNotFound)
			return
		}

		// get user
		var user _GetUser
		sql := `SELECT id, username, name, email, role, teams_ids FROM users WHERE id = $1`
		err = api.PgConn.QueryRow(c.Request.Context(), sql, id).Scan(
			&user.Id,
			&user.Username,
			&user.Name,
			&user.Email,
			&user.Role,
			&user.TeamsIds,
		)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			log.Printf("\nfail to query user, err: %s", err)
			return
		}

		c.JSON(http.StatusOK, user)
	}
}

// Get multiple users in database
// Params:
//   - "limit" Limit of users returned. Default is 30, max is 30, min is 0.
//   - "offset" Offset for searching. Default is 0, min is 0.
//
// Responses:
//   - 400 If invalid params
//   - 200 If succeeded, containing the users
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

		// search users
		sql := `SELECT id, username, name FROM users LIMIT $1 OFFSET $2`
		rows, err := api.PgConn.Query(c.Request.Context(), sql, limit, offset)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			log.Printf("\nfail to query users, err: %s", err)
			return
		}
		defer rows.Close()

		// scan users
		users := []_MGetUser{}
		for rows.Next() {
			var u _MGetUser
			rows.Scan(&u.Id, &u.Username, &u.Name)
			users = append(users, u)
		}

		c.JSON(http.StatusOK, users)
	}
}
