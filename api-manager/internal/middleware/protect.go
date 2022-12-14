package middleware

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/auth"
	"github.com/fernandotsda/nemesys/api-manager/internal/roles"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/gin-gonic/gin"
)

// APIKeyHeader is the API Key header.
const APIKeyHeader = "X-API-Key"

func validateClient(api *api.API, c *gin.Context) (meta auth.SessionMeta, err error) {
	ctx := c.Request.Context()
	sess, err := c.Cookie(auth.SessionCookieName)
	if err != nil {
		apikey := c.GetHeader(APIKeyHeader)
		if apikey == "" {
			return meta, err
		}
		apikeyMeta, err := api.Auth.ValidateAPIKey(ctx, apikey)
		if err != nil {
			return meta, err
		}
		meta.Role = apikeyMeta.Role
		meta.UserId = apikeyMeta.UserId
		return meta, nil
	}
	meta, err = api.Auth.Validate(ctx, sess)
	if err != nil {
		return meta, err
	}
	return meta, nil
}

// Protect validates the user session and role. If succeeded, save session metada in context.
// Responses:
//   - 401 If no session cookie
//   - 401 If session invalid
//   - 403 If invalid role
func Protect(api *api.API, accessLevel roles.Role) func(c *gin.Context) {
	return func(c *gin.Context) {
		meta, err := validateClient(api, c)
		if err != nil {
			if c.Request.Context().Err() != nil {
				return
			}
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if meta.Role < accessLevel {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
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
		id, err := strconv.ParseInt(c.Param("userId"), 10, 32)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}

		meta, err := validateClient(api, c)
		if err != nil {
			if c.Request.Context().Err() != nil {
				return
			}
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if meta.Role < accessLevel && meta.UserId != int32(id) {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		// save session metadata
		c.Set("sess_meta", meta)
	}
}
