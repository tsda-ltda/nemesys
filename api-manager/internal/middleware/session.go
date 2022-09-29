package middleware

import (
	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/auth"
	"github.com/fernandotsda/nemesys/api-manager/internal/uauth"
	"github.com/gin-gonic/gin"
)

func validateSession(api *api.API, c *gin.Context) (meta auth.SessionMeta, err error) {
	// get session cookie
	sess, err := c.Cookie(uauth.SessionCookieName)
	if err != nil {
		return meta, err
	}

	// validate session
	meta, err = api.Auth.Validate(c.Request.Context(), sess)
	if err != nil {
		return meta, err
	}
	return meta, nil
}
