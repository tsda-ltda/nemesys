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

// Creates a new team on databse.
// Responses:
//   - 400 If invalid body.
//   - 400 If json fields are invalid.
//   - 400 If ident is already in use.
//   - 400 If ident can be parsed to number.
//   - 200 If succeeded.
func CreateHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// bind team
		var team models.Team
		err := c.ShouldBind(&team)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		// validate team
		err = api.Validate.Struct(team)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		// validate ident
		_, err = strconv.ParseInt(team.Ident, 10, 64)
		if err == nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgIdentIsNumber))
			return
		}

		// get ident existence
		e, err := api.PgConn.Teams.ExistsIdent(ctx, team.Ident)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to check if team ident exists", logger.ErrField(err))
			return
		}

		// check if ident exists
		if e {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgIdentExists))
			return
		}

		// save team in database
		err = api.PgConn.Teams.Create(ctx, models.Team{
			Name:  team.Name,
			Ident: team.Ident,
			Descr: team.Descr,
		})
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to create team", logger.ErrField(err))
			return
		}
		api.Log.Debug("new team created, ident: " + team.Ident)

		c.Status(http.StatusOK)
	}
}
