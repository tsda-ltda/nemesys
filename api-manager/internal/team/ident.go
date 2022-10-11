package team

import (
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/gin-gonic/gin"
)

// If context 'ident' param is a valid int, return it as int. Otherwise
// Query on database for id. Returns an error if fail to query a row on database.
func getId(api *api.API, c *gin.Context) (int, error) {
	ctx := c.Request.Context()
	ident := c.Param("ident")
	// get ident as a number
	id, err := strconv.Atoi(ident)
	if err == nil {
		return id, nil
	}

	// get id from database
	sql := `SELECT id FROM teams WHERE ident = $1`
	err = api.PgConn.QueryRow(ctx, sql, ident).Scan(&id)
	return id, err
}
