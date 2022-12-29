package endpoint

import (
	"net/http"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/gin-gonic/gin"
)

// Creates a new notification endpoint.
// Responses:
//   - 400 If invalid body.
//   - 400 If json fields are invalid.
//   - 200 If succeeded.
func CreateHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		var endpoint models.AlarmEndpoint
		err := c.ShouldBind(&endpoint)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}

		err = api.Validate.Struct(endpoint)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}

		id, err := api.PG.CreateAlarmEndpoint(ctx, endpoint)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to create notification endpoint", logger.ErrField(err))
			return
		}

		c.JSON(http.StatusOK, tools.IdRes(int64(id)))
	}
}
