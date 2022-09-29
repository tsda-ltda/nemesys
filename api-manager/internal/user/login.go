package user

import (
	"log"
	"net/http"
	"os"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/auth"
	"github.com/fernandotsda/nemesys/api-manager/internal/roles"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/gin-gonic/gin"
)

const SessionCookieName = "sess"

type _Login struct {
	Username string `json:"username" validate:"required,min=2,max=50"`
	Password string `json:"password" validate:"required,min=5,max=50"`
}

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
			log.Printf("fail to get user's password, err: %s", err)
			return
		}

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
			log.Printf("fail to create user session, err: %s", err)
			return
		}

		// save session token in cookie
		c.SetCookie(SessionCookieName, token, 1000, "/", os.Getenv("HOST"), false, true)
		c.Status(http.StatusOK)
	}
}
