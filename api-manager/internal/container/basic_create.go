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
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidBody))
			return
		}

		err = api.Validate.Struct(container)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidJSONFields))
			return
		}

		container.Base.Type = types.CTBasic

		_, err = api.PG.CreateBasicContainer(ctx, container)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to create basic container", logger.ErrField(err))
			return
		}

		c.Status(http.StatusOK)
	}
}
