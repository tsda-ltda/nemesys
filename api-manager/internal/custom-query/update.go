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

		// get id
		id, err := strconv.ParseInt(c.Param("id"), 0, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		// bind body
		var cq models.CustomQuery
		err = c.ShouldBind(&cq)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		// validate body
		err = api.Validate.Struct(cq)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidJSONFields))
			return
		}

		// check if ident can be parsed to int
		_, err = strconv.Atoi(cq.Ident)
		if err == nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidJSONFields))
			return
		}

		// assign id
		cq.Id = int32(id)

		// get ident existence
		r, err := api.PgConn.CustomQueries.ExistsIdent(ctx, cq.Id, cq.Ident)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to check custom query ident existence", logger.ErrField(err))
			return
		}

		// check if exists
		if !r.Exists {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgCustomQueryNotFound))
			return
		}

		// check if ident exists
		if r.IdentExists {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgIdentExists))
			return
		}

		// update custom query
		_, err = api.PgConn.CustomQueries.Update(ctx, cq)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to update custom query", logger.ErrField(err))
			return
		}
		api.Log.Debug("custom query updated, ident: " + cq.Ident)

		c.Status(http.StatusOK)
	}
}
