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

		var team models.Team
		err := c.ShouldBind(&team)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}

		err = api.Validate.Struct(team)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}

		_, err = strconv.ParseInt(team.Ident, 10, 64)
		if err == nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgIdentIsNumber))
			return
		}

		exists, err := api.PG.TeamIdentExists(ctx, team.Ident, -1)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to check if team ident exists", logger.ErrField(err))
			return
		}
		if exists {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgIdentExists))
			return
		}

		id, err := api.PG.CreateTeam(ctx, models.Team{
			Name:  team.Name,
			Ident: team.Ident,
			Descr: team.Descr,
		})
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to create team", logger.ErrField(err))
			return
		}
		api.Log.Info("Team created, id: " + strconv.FormatInt(int64(id), 10))

		c.JSON(http.StatusOK, tools.IdRes(int64(id)))
	}
}
