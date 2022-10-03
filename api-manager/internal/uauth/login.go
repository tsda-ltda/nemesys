package uauth

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/auth"
	"github.com/fernandotsda/nemesys/api-manager/internal/roles"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/env"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/gin-gonic/gin"
)

const SessionCookieName = "sess"

type _Login struct {
	Username string `json:"username" validate:"required,min=2,max=50"`
	Password string `json:"password" validate:"required,min=5,max=50"`
}

// Login into a user account.
// Responses:
//   - 400 If invalid body.
//   - 400 If invalid body fields.
//   - 404 If username or password is incorrect.
//   - 200 If succeeded.
func LoginHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		var form _Login

		// bind login form
		err := c.ShouldBind(&form)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// validate
		err = api.Validate.Struct(form)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// get hashed password
		sql := `SELECT password, role, id FROM users WHERE username = $1`
		rows, err := api.PgConn.Query(c.Request.Context(), sql, form.Username)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to get user's password", logger.ErrField(err))
			return
		}
		defer rows.Close()

		// scan rows
		var hashedPw string
		var id int
		var role roles.Role
		for rows.Next() {
			rows.Scan(&hashedPw, &role, &id)
		}

		// check if user exists and password is correct
		if rows.CommandTag().RowsAffected() == 0 || !auth.CheckHash(form.Password, hashedPw) {
			c.JSON(http.StatusUnauthorized, tools.NewMsg("username or password is wrong"))
			return
		}

		// create new session
		token, err := api.Auth.NewSession(c.Request.Context(), auth.SessionMeta{
			UserId: id,
			Role:   role,
		})
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to create user session", logger.ErrField(err))
			return
		}

		// save session token in cookie
		ttl, _ := strconv.Atoi(env.UserSessionTTL)
		c.SetCookie(SessionCookieName, token, ttl, "/", env.APIManagerHost, false, true)

		c.Status(http.StatusOK)
	}
}
