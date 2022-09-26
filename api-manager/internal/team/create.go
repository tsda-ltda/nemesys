package team

import (
	"log"
	"net/http"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/gin-gonic/gin"
)

// Team struct for CreateHandler json requests.
type _CreateTeam struct {
	Name  string `json:"name" validate:"required,max=50,min=2"`
	Ident string `json:"ident" validate:"required,max=50,min=2"`
}

// Creates a new team on databse.
// Responses:
//   - 400 If invalid body.
//   - 400 If json fields are invalid.
//   - 400 If ident is already in use.
//   - 200 If succeeded.
func CreateHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		var team _CreateTeam

		// bind team
		err := c.ShouldBind(&team)
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

		// check if ident exists in database
		var identInUse bool
		sql := `SELECT EXISTS (
				SELECT 1 FROM teams WHERE ident = $1
			);
		`

		// query row
		err = api.PgConn.QueryRow(c.Request.Context(), sql, team.Ident).Scan(&identInUse)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			log.Printf("\nfail to query team's ident, err: %s", err)
			return
		}

		// check if ident is in use
		if identInUse {
			c.JSON(http.StatusBadRequest, tools.NewMsg("ident already in use"))
			return
		}

		// save team in database
		sql = `INSERT INTO teams (name, ident, users_ids) VALUES($1, $2, $3)`
		_, err = api.PgConn.Exec(c.Request.Context(), sql, team.Name, team.Ident, []int{})
		if err != nil {
			c.Status(http.StatusInternalServerError)
			log.Printf("\nfail to create team, err: %s", err)
			return
		}
		log.Printf("\nteam '%s' created successfuly", team.Ident)

		c.Status(http.StatusOK)
	}
}
