package middleware

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/roles"
	"github.com/gin-gonic/gin"
)

// Protect validates the user session and role. If succeeded, save session metada in context.
// Responses:
//   - 401 If no session cookie
//   - 401 If session invalid
//   - 403 If invalid role
func Protect(api *api.API, accessLevel roles.Role) func(c *gin.Context) {
	return func(c *gin.Context) {
		// validate session
		meta, err := validateSession(api, c)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// validate role
		if meta.Role < accessLevel {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		// save session metadata
		c.Set("sess_meta", meta)
	}
}

// Protect validates the user session and role. Allowing users with required roles,
// or if user info belogs to user. If succeeded, save session metada in context.
// Responses:
//   - 400 If invalid id
//   - 401 If no session cookie
//   - 401 If session invalid
//   - 403 If invalid role
func ProtectUser(api *api.API, accessLevel roles.Role) func(c *gin.Context) {
	return func(c *gin.Context) {
		// get user id
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		// validate session
		meta, err := validateSession(api, c)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// validate role
		if meta.Role < accessLevel && meta.UserId != id {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		// save session metadata
		c.Set("sess_meta", meta)
	}
}
