package trap

import (
	"net/http"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/gin-gonic/gin"
)

// Creates a trap listener.
// Responses:
//   - 400 If invalid body.
//   - 400 If json fields are invalid.
//   - 400 If host and port exists.
//   - 200 If succeeded.
func CreateHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		var tl models.TrapListener
		err := c.ShouldBind(&tl)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}

		err = api.Validate.Struct(tl)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidJSONFields))
			return
		}

		exists, err := api.PG.AlarmCategoryExists(ctx, tl.AlarmCategoryId)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to check if alarm category exists", logger.ErrField(err))
			return
		}
		if !exists {
			c.JSON(http.StatusNotFound, tools.MsgRes(tools.MsgAlarmCategoryNotFound))
			return
		}

		exists, err = api.PG.TrapListenerHostPortExists(ctx, tl.Host, tl.Port, -1)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to check if trap listener host and port exists", logger.ErrField(err))
			return
		}
		if exists {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgTrapListerHostPortExists))
			return
		}

		id, err := api.PG.CreateTrapListener(ctx, tl)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to create trap listener", logger.ErrField(err))
			return
		}
		tl.Id = id
		api.CreateTrapListener(tl)

		api.Log.Debug("New trap listener created")

		c.JSON(http.StatusOK, tools.EmptyRes())
	}
}
