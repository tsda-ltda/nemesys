package user

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/gin-gonic/gin"
)

// Get user in database
// Responses:
//   - 400 If invalid id
//   - 404 If user not foud
//   - 200 If succeeded
func GetHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// get id from param
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// get user
		user, e, err := api.PgConn.Users.GetWithoutPW(ctx, id)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to check if user exists", logger.ErrField(err))
			return
		}

		// check if user exists
		if !e {
			c.Status(http.StatusNotFound)
			return
		}

		c.JSON(http.StatusOK, user)
	}
}

// Get multiple users in database.
// Params:
//   - "limit" Limit of users returned. Default is 30, max is 30, min is 0.
//   - "offset" Offset for searching. Default is 0, min is 0.
//
// Responses:
//   - 400 If invalid params.
//   - 200 If succeeded.
func MGetHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// get limit
		limit, err := tools.IntRangeQuery(c, "limit", 30, 30, 1)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// get offset
		offset, err := tools.IntMinQuery(c, "offset", 0, 0)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// get users
		users, err := api.PgConn.Users.MGetSimplified(ctx, limit, offset)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to query users", logger.ErrField(err))
			return
		}

		c.JSON(http.StatusOK, users)
	}
}
