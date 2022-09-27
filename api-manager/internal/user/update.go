package user

import (
	"log"
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/gin-gonic/gin"
)

// Creates a new user on database.
// Responses:
//   - 400 If invalid body.
//   - 400 If json fields are invalid.
//   - 400 If username or email already in use.
//   - 404 If user not founded.
//   - 200 If succeeded.
func UpdateHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		// get id param
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.Status((http.StatusBadRequest))
			return
		}

		// bind user
		var user _CreateUser
		err = c.ShouldBind(&user)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// validate user
		err = api.Validate.Struct(user)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// set id
		user.Id = id

		var exists, usernameInUse, emailInUse bool
		sql := `SELECT EXISTS (
				SELECT 1 FROM users WHERE id = $1
			) as EX1, EXISTS (
				SELECT 1 FROM users WHERE id != $1 AND username = $2
			) as EX2, EXISTS (
				SELECT 1 FROM users WHERE id != $1 AND email = $3
			) as EX3
		`
		err = api.PgConn.QueryRow(c.Request.Context(), sql, id, user.Username, user.Email).Scan(
			&exists, &usernameInUse, &emailInUse,
		)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			log.Printf("\nfail to query in users, err: %s", err)
			return
		}

		// check if user exists
		if !exists {
			c.Status(http.StatusNotFound)
			return
		}

		// check if username is in use
		if usernameInUse {
			c.JSON(http.StatusBadRequest, tools.NewMsg("username already in use"))
			return
		}

		// check if email is in use
		if emailInUse {
			c.JSON(http.StatusBadRequest, tools.NewMsg("email already in use"))
			return
		}

		// update user
		sql = `UPDATE users SET 
		(name, username, password, email, role) =
		($1, $2, $3, $4, $5) WHERE id = $6
		`
		_, err = api.PgConn.Exec(c.Request.Context(), sql,
			&user.Name,
			&user.Username,
			&user.Password,
			&user.Email,
			&user.Role,
			id,
		)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			log.Printf("\nfail to update user, err: %s", err)
			return
		}

		c.Status(http.StatusOK)
	}
}