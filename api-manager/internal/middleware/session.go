package middleware

import (
	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/auth"
	"github.com/gin-gonic/gin"
)

func validateSession(api *api.API, c *gin.Context) (meta auth.SessionMeta, err error) {
	sess, err := c.Cookie(auth.SessionCookieName)
	if err != nil {
		return meta, err
	}
	meta, err = api.Auth.Validate(c.Request.Context(), sess)
	if err != nil {
		return meta, err
	}
	return meta, nil
}
