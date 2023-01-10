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

		rawId := c.Param("ctxId")
		ctxId, err := strconv.ParseInt(rawId, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}

		teamId, err := strconv.ParseInt(c.Param("teamId"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}

		var context models.Context
		err = c.ShouldBind(&context)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidBody))
			return
		}

		err = api.Validate.Struct(context)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidJSONFields))
			return
		}

		_, err = strconv.ParseInt(context.Ident, 10, 64)
		if err == nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgIdentIsNumber))
			return
		}

		teamExists, identExists, err := api.PG.ExistsTeamAndContextIdent(ctx, int32(teamId), context.Ident, int32(ctxId))
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			api.Log.Error("Fail to get context existence", logger.ErrField(err))
			c.Status(http.StatusInternalServerError)
			return
		}
		if !teamExists {
			c.JSON(http.StatusNotFound, tools.MsgRes(tools.MsgTeamNotFound))
			return
		}
		if identExists {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgIdentExists))
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

		api.Log.Info("Context created with success, id: " + rawId)
		c.JSON(http.StatusOK, tools.EmptyRes())
	}
}
