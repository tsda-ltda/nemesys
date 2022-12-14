package container

import (
	"net/http"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/types"
	"github.com/gin-gonic/gin"
)

// Creates a basic container.
// Responses:
//   - 400 If invalid body.
//   - 400 If json fields are invalid.
//   - 200 If succeeded.
func CreateBasicHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		var container models.Container[struct{}]
		err := c.ShouldBind(&container)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidBody))
			return
		}

		err = api.Validate.Struct(container)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidJSONFields))
			return
		}

		container.Base.Type = types.CTBasic

		id, err := api.PG.CreateBasicContainer(ctx, container)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to create basic container", logger.ErrField(err))
			return
		}

		c.JSON(http.StatusOK, tools.IdRes(int64(id)))
	}
}
