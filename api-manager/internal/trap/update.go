package trap

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/gin-gonic/gin"
)

// Updates a trap listener.
// Responses:
//   - 400 If invalid body.
//   - 400 If json fields are invalid.
//   - 400 If host and port exists.
//   - 404 If not found.
//   - 200 If succeeded.
func UpdateHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		rawId := c.Param("listenerId")
		id, err := strconv.ParseInt(rawId, 0, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		var tl models.TrapListener
		err = c.ShouldBind(&tl)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		err = api.Validate.Struct(tl)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidJSONFields))
			return
		}

		tl.Id = int32(id)

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
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgAlarmCategoryNotFound))
			return
		}

		exists, err = api.PG.TrapListenerHostPortExists(ctx, tl.Host, tl.Port, int32(id))
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to check if trap listener host and port exists", logger.ErrField(err))
			return
		}
		if exists {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgTrapListerHostPortExists))
			return
		}

		exists, err = api.PG.UpdateTrapListener(ctx, tl)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to update trap listener", logger.ErrField(err))
			return
		}
		if !exists {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgTrapListenerNotFound))
			return
		}
		api.UpdateTrapListener(tl)

		c.Status(http.StatusOK)
	}
}
