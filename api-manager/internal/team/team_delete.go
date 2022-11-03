package team

import (
	"fmt"
	"net/http"
	"strconv"

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
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// delete team
		e, err := api.PgConn.Teams.Delete(ctx, int32(id))
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to delete team", logger.ErrField(err))
			return
		}

		// check if team existed
		if !e {
			c.Status(http.StatusNotFound)
			return
		}
		api.Log.Debug("team deleted, id: " + fmt.Sprint(id))

		c.Status(http.StatusNoContent)
	}
}
