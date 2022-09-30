package uauth

import (
	"log"
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/roles"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/gin-gonic/gin"
)

// Logout of a user account.
// Responses:
//   - 400 If no session was running.
//   - 200 If succeeded.
//
// Keys dependencies:
//   - "sess_meta" Session metadata.
func Logout(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		meta, err := tools.GetSessionMeta(c)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			log.Println(err)
			return
		}

		// remove session
		err = api.Auth.RemoveSession(c.Request.Context(), meta.UserId)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		c.Status(http.StatusOK)
	}
}

// Force a user logout.
// Responses:
//   - 400 If invalid id.
//   - 400 If no session was running.
//   - 200 If succeeded.
func ForceLogout(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		// get session metadata
		meta, err := tools.GetSessionMeta(c)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			log.Println(err)
			return
		}

		// get userId
		userId, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// get target user's role
		sql := `SELECT role FROM users WHERE id = $1`
		rows, err := api.PgConn.Query(c.Request.Context(), sql, userId)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			log.Printf("fail to query user role, err: %s", err)
			return
		}
		defer rows.Close()

		// scan role
		var userRole roles.Role
		for rows.Next() {
			rows.Scan(&userRole)
		}

		// check if user does not exists
		if rows.CommandTag().RowsAffected() == 0 {
			c.Status(http.StatusBadRequest)
			return
		}

		// check if target's role is superior
		if userRole > meta.Role {
			c.Status(http.StatusForbidden)
			return
		}

		// remove session
		err = api.Auth.RemoveSession(c.Request.Context(), userId)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		c.Status(http.StatusOK)
	}
}
