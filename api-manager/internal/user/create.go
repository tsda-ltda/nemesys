package user

import (
	"log"
	"net/http"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/gin-gonic/gin"
)

// User struct for CreateHandler json responses.
type _CreateUser struct {
	Id       int    `json:"id" validate:"-"`
	Role     int    `json:"role" validate:"required"`
	Name     string `json:"name" validate:"required,min=2,max=50"`
	Username string `json:"username" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=5,max=50"`
	Email    string `json:"email" validate:"required,email"`
}

// Creates a new user on database.
// Responses:
//   - 400 If invalid body.
//   - 400 If json fields are invalid.
//   - 400 If username is already in use.
//   - 400 If email is already in use.
//   - 200 If succeeded.
func CreateHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		var user _CreateUser

		// bind user
		err := c.ShouldBind(&user)
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

		// check if username and email exists in database
		var usernameInUse, emailInUse bool
		sql := `SELECT EXISTS (
				SELECT 1 FROM users WHERE username = $1
			) as EX1, EXISTS (
				SELECT 1 FROM users WHERE email = $2
			) as EX2;
		`

		// query row
		err = api.PgConn.QueryRow(c.Request.Context(), sql, user.Username, user.Email).Scan(&usernameInUse, &emailInUse)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			log.Printf("\nfail to query username and email, err: %s", err)
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

		// save user in database
		sql = `INSERT INTO users (name, username, password, email, role)
		VALUES($1, $2, $3, $4, $5)`
		_, err = api.PgConn.Exec(c.Request.Context(), sql,
			user.Name,
			user.Username,
			user.Password,
			user.Email,
			user.Role,
		)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			log.Printf("\nfail to create user, err: %s", err)
			return
		}
		log.Printf("\nuser '%s' created successfuly", user.Username)

		c.Status(http.StatusOK)
	}
}
