package user

import (
	"log"
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/gin-gonic/gin"
)

// Creates a new user on databse
// Responses:
//   - 400 If invalid body
//   - 400 If user's fields are invalid
//   - 404 If user not founded
//   - 200 If succeeded
func UpdateHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		// get id param
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.Status((http.StatusBadRequest))
			return
		}

		// bind user
		var user models.User
		err = c.ShouldBind(&user)
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

		// set id
		user.Id = id

		// check if user exists
		e, err := api.PgConn.Users.Exists(c.Request.Context(), id)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			log.Printf("\nfail to query user, err: %s", err)
			return
		}
		if !e {
			c.Status(http.StatusNotFound)
			return
		}

		// update user
		_, err = api.PgConn.Users.Update(c.Request.Context(), user)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			log.Printf("\nfail to update user, err: %s", err)
			return
		}

		c.Status(http.StatusOK)
	}
}
