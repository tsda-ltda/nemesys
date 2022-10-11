package team

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/gin-gonic/gin"
)

// Team struct for CreateHandler json requests
type _UpdateTeam struct {
	Name  string `json:"name" validate:"required,max=50,min=2"`
	Ident string `json:"ident" validate:"required,max=50,min=2"`
	Descr string `json:"descr" validate:"max=255"`
}

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
		id, err := getId(api, c)
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}

		// bind team
		var team _UpdateTeam
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
		if team.Ident != c.Param("ident") {
			var identInUse bool
			sql := `SELECT EXISTS (SELECT 1 FROM teams WHERE ident = $1);`

			// query row
			err = api.PgConn.QueryRow(ctx, sql, team.Ident).Scan(&identInUse)
			if err != nil {
				c.Status(http.StatusInternalServerError)
				api.Log.Error("fail to query team by ident", logger.ErrField(err))
				return
			}

			if identInUse {
				c.JSON(http.StatusBadRequest, tools.NewMsg("ident already in use"))
				return
			}
		}

		// update team in database
		sql := `UPDATE teams SET (name, ident, descr) = ($1, $2, $3) WHERE id = $4`
		f, err := api.PgConn.Exec(ctx, sql, team.Name, team.Ident, team.Descr, id)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to update team", logger.ErrField(err))
			return
		}

		// check if team exists
		if f.RowsAffected() == 0 {
			c.Status(http.StatusNotFound)
			return
		}

		api.Log.Debug(fmt.Sprintf("team '%s' updated successfully", team.Ident))

		c.Status(http.StatusOK)
	}
}
