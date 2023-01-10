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

// Updates a team on database.
// Responses:
//   - 400 If invalid body.
//   - 400 If json fields are invalid
//   - 400 If ident is already in use.
//   - 400 If ident can be parsed to number.
//   - 404 If team does not exists.
//   - 200 If succeeded.
func UpdateHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		rawId := c.Param("teamId")
		id, err := strconv.ParseInt(rawId, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}

		var team models.Team
		err = c.ShouldBind(&team)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidBody))
			return
		}

		err = api.Validate.Struct(team)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidJSONFields))
			return
		}

		_, err = strconv.ParseInt(team.Ident, 10, 64)
		if err == nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgIdentIsNumber))
			return
		}

		exists, err := api.PG.TeamIdentExists(ctx, team.Ident, int32(id))
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to check if ident is available", logger.ErrField(err))
			return
		}
		if exists {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgIdentExists))
			return
		}

		team.Id = int32(id)
		exists, err = api.PG.UpdateTeam(ctx, team)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to update team", logger.ErrField(err))
			return
		}
		if !exists {
			c.JSON(http.StatusNotFound, tools.MsgRes(tools.MsgTeamNotFound))
			return
		}

		api.Log.Info("Team updated, id: " + rawId)

		c.JSON(http.StatusOK, tools.EmptyRes())
	}
}
