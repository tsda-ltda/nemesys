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

// Creates a new contextual metric.
// Responses:
//   - 400 If invalid body.
//   - 400 If json fields are invalid.
//   - 404 If context or metric does not exists.
//   - 400 If ident is already in use.
//   - 200 If succeeded.
func CreateHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// context id
		ucontextId, err := strconv.Atoi(c.Param("ctxId"))
		if err != nil {
			c.Status(http.StatusBadGateway)
			return
		}
		contextId := int32(ucontextId)

		// bind body
		var cmetric models.ContextualMetric
		err = c.ShouldBind(&cmetric)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// validate contextual metric
		err = api.Validate.Struct(cmetric)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		ce, me, ie, err := api.PgConn.ContextualMetrics.ExistsContextMetricAndIdent(ctx, contextId, cmetric.MetricId, cmetric.Ident, -1)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to check if context, metric and ident exists", logger.ErrField(err))
			return
		}

		// check if context exists
		if !ce {
			c.Status(http.StatusNotFound)
			return
		}

		// check if metric exists
		if !me {
			c.Status(http.StatusNotFound)
			return
		}

		// check if ident exists
		if ie {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgIdentExists))
			return
		}

		// assign context id
		cmetric.ContextId = contextId

		// create contextual metric
		err = api.PgConn.ContextualMetrics.Create(ctx, cmetric)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		api.Log.Debug("contexual metric created, ident: " + cmetric.Ident)

		c.Status(http.StatusOK)
	}
}
