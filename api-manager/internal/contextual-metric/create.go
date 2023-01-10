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

		contextId, err := strconv.ParseInt(c.Param("ctxId"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}

		var cmetric models.ContextualMetric
		err = c.ShouldBind(&cmetric)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidBody))
			return
		}

		err = api.Validate.Struct(cmetric)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidJSONFields))
			return
		}

		cmetric.ContextId = int32(contextId)

		r, err := api.PG.ContextMetricAndContexualMetricIdentExists(ctx, cmetric.ContextId, cmetric.MetricId, cmetric.Ident, -1)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to check if context, metric and ident exists", logger.ErrField(err))
			return
		}

		if !r.ContextExists {
			c.JSON(http.StatusNotFound, tools.MsgRes(tools.MsgContextNotFound))
			return
		}
		if !r.MetricExists {
			c.JSON(http.StatusNotFound, tools.MsgRes(tools.MsgMetricNotFound))
			return
		}
		if r.IdentExists {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgIdentExists))
			return
		}

		id, err := api.PG.CreateContextualMetric(ctx, cmetric)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			return
		}
		api.Log.Debug("Contextual metric created, id: " + strconv.FormatInt(id, 10))

		c.JSON(http.StatusOK, tools.IdRes(id))
	}
}
