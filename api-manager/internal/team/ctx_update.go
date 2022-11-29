package team

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/gin-gonic/gin"
)

// Updates a context.
// Responses:
//   - 400 If invalid body.
//   - 400 If json fields are invalid.
//   - 400 If ident is already in use.
//   - 400 If ident can be parsed to number.
//   - 400 If invalid id.
//   - 404 If not found.
//   - 200 If succeeded.
func UpdateContextHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		ctxId, err := strconv.ParseInt(c.Param("ctxId"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		teamId, err := strconv.ParseInt(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		var context models.Context
		err = c.ShouldBind(&context)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidBody))
			return
		}

		err = api.Validate.Struct(context)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidJSONFields))
			return
		}

		_, err = strconv.ParseInt(context.Ident, 10, 64)
		if err == nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgIdentIsNumber))
			return
		}

		r, err := api.PG.ExistsTeamAndContextIdent(ctx, int32(teamId), context.Ident, int32(ctxId))
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			api.Log.Error("Fail to get context existence", logger.ErrField(err))
			c.Status(http.StatusInternalServerError)
			return
		}
		if !r.TeamExists {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgTeamNotFound))
			return
		}
		if r.IdentExists {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgIdentExists))
			return
		}

		_, err = api.PG.UpdateContext(ctx, models.Context{
			Name:  context.Name,
			Ident: context.Ident,
			Descr: context.Descr,
			Id:    int32(ctxId),
		})
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			api.Log.Error("Fail to create context", logger.ErrField(err))
			c.Status(http.StatusInternalServerError)
			return
		}

		api.Log.Debug("Context created with success, ident: " + context.Ident)
		c.Status(http.StatusOK)
	}
}
