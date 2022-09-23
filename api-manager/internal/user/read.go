package user

import (
	"net/http"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/gin-gonic/gin"
)

// User struct for MGetHandler responses
type _MGetUser struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
}

// Creates a new user on databse
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
