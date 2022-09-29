package tools

import (
	"fmt"

	"github.com/fernandotsda/nemesys/api-manager/internal/auth"
	"github.com/gin-gonic/gin"
)

// GetSessionMeta returns the session metadata saved in the request context.
// If fail to get session metadata or to make type assertion, returns an error.
func GetSessionMeta(c *gin.Context) (meta auth.SessionMeta, err error) {
	// get session metadata
	_metadata, e := c.Get("sess_meta")
	if !e {
		return meta, fmt.Errorf("fail to get metadata in gin context keys")
	}

	// type assertion
	meta, ok := _metadata.(auth.SessionMeta)
	if !ok {
		return meta, fmt.Errorf("fail to make type assertion into session metadata, meta: %v", meta)
	}
	return meta, nil
}
