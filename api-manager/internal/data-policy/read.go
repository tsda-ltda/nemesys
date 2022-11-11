package datapolicy

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/gin-gonic/gin"
)

// Get all data policies.
// Responses:
//   - 200 If succeeded.
func MGetHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// get data policies
		dps, err := api.PgConn.DataPolicy.MGet(ctx)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to read data policies", logger.ErrField(err))
			return
		}
		c.JSON(http.StatusOK, dps)
	}
}

// Get a data policy.
// Responses:
//   - 400 If invalid params.
//   - 404 If not found.
//   - 200 If succeeded.
func GetHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// get data policy id
		id, err := strconv.ParseInt(c.Param("id"), 10, 16)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
		}

		// get data policies
		r, err := api.PgConn.DataPolicy.Get(ctx, int16(id))
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to read data policy", logger.ErrField(err))
			return
		}

		// check if exists
		if !r.Exists {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgDataPolicyNotFound))
			return
		}

		c.JSON(http.StatusOK, r.DataPolicy)
	}
}
