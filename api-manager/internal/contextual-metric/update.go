package ctxmetric

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/gin-gonic/gin"
)

// Updates a contextual metric.
// Responses:
//   - 400 If invalid id.
//   - 400 If invalid body.
//   - 400 If invalid json fields.
//   - 404 If contextual metric does not exists.
//   - 204 If succeeded.
func UpdateHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// context id
		rawCtxId := c.Param("ctxId")
		ctxId, err := strconv.ParseInt(rawCtxId, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		// contextual metric id
		rawId := c.Param("metricId")
		id, err := strconv.ParseInt(rawId, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		// bind body
		var cmetric models.ContextualMetric
		err = c.ShouldBind(&cmetric)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidBody))
			return
		}

		// validate body
		err = api.Validate.Struct(cmetric)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidJSONFields))
			return
		}

		// check if ident is in use
		ie, err := api.PgConn.ContextualMetrics.ExistsIdent(ctx, cmetric.Ident, int32(ctxId), int64(id))
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to check if contextual metric exits", logger.ErrField(err))
			return
		}
		if ie {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgIdentExists))
			return
		}

		// assign id
		cmetric.Id = int64(id)

		// update
		e, err := api.PgConn.ContextualMetrics.Update(ctx, cmetric)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to update contextual metric", logger.ErrField(err))
			return
		}

		// check if exists
		if !e {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgContextualMetricNotFound))
			return
		}
		api.Log.Debug("contextual metric updated, id: " + rawId)

		c.Status(http.StatusOK)
	}
}
