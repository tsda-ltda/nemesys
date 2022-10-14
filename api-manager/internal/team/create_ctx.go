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

const msgContextIdentExists = "Ident already exists."

// Creates a new context.
// Responses:
//   - 400 If invalid body.
//   - 400 If json fields are invalid.
//   - 400 If ident is already in use.
//   - 200 If succeeded.
func CreateContextHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// get team id
		teamId, err := strconv.Atoi(c.Param("teamId"))
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// bind context
		var context models.ContextCreateReq
		err = c.ShouldBind(&context)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// validate struct
		err = api.Validate.Struct(context)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// get ident existence
		e, err := api.PgConn.Contexts.ExistsIdent(ctx, context.Ident)
		if err != nil {
			api.Log.Error("fail to get context existence", logger.ErrField(err))
			c.Status(http.StatusInternalServerError)
			return
		}

		// check if ident exists
		if e {
			c.JSON(http.StatusBadGateway, tools.JSONMSG(msgContextIdentExists))
			return
		}

		// create context
		err = api.PgConn.Contexts.Create(ctx, models.Context{
			TeamId: teamId,
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
