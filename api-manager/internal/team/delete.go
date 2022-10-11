package team

import (
	"net/http"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/gin-gonic/gin"
)

// Deletes team from databse
// Responses:
//   - 404 If team not founded
//   - 204 If succeeded
func DeleteHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		// get id
		id, err := getId(api, c)
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}

		// delete team
		t, err := api.PgConn.Exec(ctx, "DELETE FROM teams WHERE id = $1", id)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to delete team", logger.ErrField(err))
			return
		}
		if t.RowsAffected() == 0 {
			c.Status(http.StatusNotFound)
			return
		}
		api.Log.Debug("team deleted, ident: " + c.Param("ident"))

		c.Status(http.StatusNoContent)
	}
}
