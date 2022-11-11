package datapolicy

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/env"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/gin-gonic/gin"
)

// Creates a new data policy.
// Responses:
//   - 400 If invalid body.
//   - 400 If json fields are invalid.
//   - 400 If exceeds the maximum number of data policies.
//   - 200 If succeeded.
func CreateHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// bind data policy
		var dp models.DataPolicy
		err := c.ShouldBind(&dp)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		// validate struct
		err = api.Validate.Struct(dp)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidBody))
			return
		}

		// get number of data policies in the system
		n, err := api.PgConn.DataPolicy.Count(ctx)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to count number of data polcies", logger.ErrField(err))
			return
		}

		// get maximum permited data policies
		max, err := strconv.ParseInt(env.MaxDataPolicies, 10, 0)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to parse env.MaxDataPolicies", logger.ErrField(err))
			return
		}

		// check if exceeds max data policies
		if n >= max {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgMaxDataPolicy))
			api.Log.Warn("attempt to create data-policy failed, maximum number reached")
			return
		}

		// create data policy
		_, err = api.PgConn.DataPolicy.Create(ctx, dp)
		if err != nil {
			api.Log.Error("fail to create new data policy", logger.ErrField(err))
			c.Status(http.StatusInternalServerError)
			return
		}
		api.Log.Info("data policy created")
		c.Status(http.StatusOK)
	}
}
