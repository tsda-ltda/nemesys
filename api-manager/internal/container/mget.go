package container

import (
	"net/http"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/types"
	"github.com/gin-gonic/gin"
)

// Get multi base containers on database
// Params:
//   - "limit" Limit of teams returned. Default is 30, max is 30, min is 0.
//   - "offset" Offset for searching. Default is 0, min is 0.
//
// Responses:
//   - 400 If invalid params.
//   - 200 If succeeded.
func MGet(api *api.API, t types.ContainerType) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// db query params
		limit, err := tools.IntRangeQuery(c, "limit", 30, 30, 1)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		offset, err := tools.IntMinQuery(c, "offset", 0, 0)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// get containers
		containers, err := api.PgConn.Containers.MGet(ctx, t, limit, offset)
		if err != nil {
			api.Log.Error("fail to get containers", logger.ErrField(err))
			c.Status(http.StatusInternalServerError)
			return
		}

		c.JSON(http.StatusOK, containers)
	}
}
