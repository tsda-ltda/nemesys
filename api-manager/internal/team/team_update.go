package team

import (
	"fmt"
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
//   - 404 If team does not exists.
//   - 200 If succeeded.
func UpdateHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// get team id
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// bind team
		var team models.Team
		err = c.ShouldBind(&team)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// validate team
		err = api.Validate.Struct(team)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// validate ident
		_, err = strconv.Atoi(team.Ident)
		if err == nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// check if ident is already in use
		e, err := api.PgConn.Teams.IdentAvailableUpdate(ctx, int32(id), team.Ident)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to check if ident is available", logger.ErrField(err))
			return
		}
		if e {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgIdentExists))
			return
		}

		// update team in database
		team.Id = int32(id)
		e, err = api.PgConn.Teams.Update(ctx, team)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to update team", logger.ErrField(err))
			return
		}

		// check if team exists
		if !e {
			c.Status(http.StatusNotFound)
			return
		}

		api.Log.Debug("team updated successfully, id" + fmt.Sprint(id))

		c.Status(http.StatusOK)
	}
}