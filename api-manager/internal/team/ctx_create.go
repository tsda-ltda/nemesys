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

// Creates a new context.
// Responses:
//   - 400 If invalid body.
//   - 400 If json fields are invalid.
//   - 400 If ident is already in use.
//   - 400 If ident can be parsed to number.
//   - 200 If succeeded.
func CreateContextHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// get team id
		teamId, err := strconv.ParseInt(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		// bind context
		var context models.Context
		err = c.ShouldBind(&context)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidBody))
			return
		}

		// validate struct
		err = api.Validate.Struct(context)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidJSONFields))
			return
		}

		// validate ident
		_, err = strconv.ParseInt(context.Ident, 10, 64)
		if err == nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgIdentIsNumber))
			return
		}

		// get ident existence
		r, err := api.PgConn.Contexts.ExistsTeamAndIdent(ctx, int32(teamId), context.Ident, -1)
		if err != nil {
			api.Log.Error("fail to get context existence", logger.ErrField(err))
			c.Status(http.StatusInternalServerError)
			return
		}

		// check if team exists
		if !r.TeamExists {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgTeamNotFound))
			return
		}

		// check if ident exists
		if r.IdentExists {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgIdentExists))
			return
		}

		// create context
		_, err = api.PgConn.Contexts.Create(ctx, models.Context{
			TeamId: int32(teamId),
			Name:   context.Name,
			Ident:  context.Ident,
			Descr:  context.Descr,
		})
		if err != nil {
			api.Log.Error("fail to create context", logger.ErrField(err))
			c.Status(http.StatusInternalServerError)
			return
		}

		api.Log.Debug("context created with success, ident: " + context.Ident)
		c.Status(http.StatusOK)
	}
}
