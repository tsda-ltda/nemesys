package customquery

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/gin-gonic/gin"
)

// Updates a custom query.
// Responses:
//   - 400 If invalid body.
//   - 400 If json fields are invalid.
//   - 404 If custom query does not exists.
//   - 400 If ident is already in use.
//   - 200 If succeeded.
func UpdateHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		id, err := strconv.ParseInt(c.Param("id"), 0, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		var cq models.CustomQuery
		err = c.ShouldBind(&cq)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		err = api.Validate.Struct(cq)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidJSONFields))
			return
		}

		_, err = strconv.Atoi(cq.Ident)
		if err == nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidJSONFields))
			return
		}

		cq.Id = int32(id)

		r, err := api.PG.ExistsCustomQueryIdent(ctx, cq.Id, cq.Ident)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to check custom query ident existence", logger.ErrField(err))
			return
		}
		if !r.Exists {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgCustomQueryNotFound))
			return
		}
		if r.IdentExists {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgIdentExists))
			return
		}

		_, err = api.PG.UpdateCustomQuery(ctx, cq)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to update custom query", logger.ErrField(err))
			return
		}
		api.Log.Debug("custom query updated, ident: " + cq.Ident)

		c.Status(http.StatusOK)
	}
}
