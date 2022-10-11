package user

import (
	"net/http"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/auth"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/gin-gonic/gin"
)

// User struct for CreateHandler json responses.
type _CreateUser struct {
	Id       int    `json:"id" validate:"-"`
	Role     int    `json:"role" validate:"required,min=1,max=4"`
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
		ctx := c.Request.Context()
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
		err = api.PgConn.QueryRow(ctx, sql, user.Username, user.Email).Scan(&usernameInUse, &emailInUse)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to check username and email on postgres", logger.ErrField(err))
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

		// hash password
		pwHashed, err := auth.Hash(user.Password, api.UserPWBcryptCost)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to hash password", logger.ErrField(err))
			return
		}

		// save user in database
		sql = `INSERT INTO users (name, username, password, email, role)
		VALUES($1, $2, $3, $4, $5)`
		_, err = api.PgConn.Exec(ctx, sql,
			user.Name,
			user.Username,
			pwHashed,
			user.Email,
			user.Role,
		)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to create user", logger.ErrField(err))
			return
		}
		api.Log.Debug("new user created, username: " + user.Username)

		c.Status(http.StatusOK)
	}
}
