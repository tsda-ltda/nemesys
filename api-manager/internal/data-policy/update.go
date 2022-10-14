package datapolicy

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/gin-gonic/gin"
)

// Update a data policy.
// Responses:
//   - 400 If invalid id.
//   - 400 If invalid body.
//   - 400 If invalid body fields.
//   - 404 If data policy not found.
//   - 200 If succeeded.
func UpdateHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// get data policy id
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// bind body
		var dp models.DataPolicy
		err = c.ShouldBind(&dp)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// validate struct
		err = api.Validate.Struct(dp)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// update data policy
		dp.Id = id
		e, err := api.PgConn.DataPolicy.Update(ctx, dp)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to update data policy", logger.ErrField(err))
			return
		}
		if !e {
			c.Status(http.StatusNotFound)
			return
		}
		api.Log.Info("data policy updated, id: " + fmt.Sprint(id))

		c.Status(http.StatusOK)
	}
}
