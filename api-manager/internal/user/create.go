package user

import (
	"log"
	"net/http"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/gin-gonic/gin"
)

// Creates a new user on databse
// Responses:
//   - 400 If invalid body
//   - 400 If user's fields are invalid
//   - 400 If username is already in use
//   - 400 If email is already in use
//   - 200 If succeeded
func CreateHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		var user models.User

		// bind user
		err := c.ShouldBind(&user)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// validate user
		err = api.Validate.Struct(user)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// Check if username is in use
		e, err := api.PgConn.Users.ExistsByUsername(c.Request.Context(), user.Username)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			log.Printf("\nfail to query username, err: %s", err)
			return
		}
		if e {
			c.JSON(http.StatusBadRequest, tools.NewMsg("username already exists"))
			return
		}

		// check if email is in use
		e, err = api.PgConn.Users.ExistsByEmail(c.Request.Context(), user.Email)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			log.Printf("\nfail to query email, err: %s", err)
			return
		}
		if e {
			c.JSON(http.StatusBadRequest, tools.NewMsg("email already exists"))
			return
		}

		// create user
		_, err = api.PgConn.Users.Create(c.Request.Context(), user)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			log.Printf("\nfail to create user, err: %s", err)
			return
		}
		log.Printf("\nuser '%s' created successfuly", user.Username)

		c.Status(http.StatusOK)
	}
}
