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
//   - 200 If succeeded.
func UpdateHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		rawCtxId := c.Param("ctxId")
		ctxId, err := strconv.ParseInt(rawCtxId, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}

		rawId := c.Param("ctxMetricId")
		id, err := strconv.ParseInt(rawId, 10, 64)
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

		ie, err := api.PG.ContextualMetricIdentExists(ctx, cmetric.Ident, int32(ctxId), int64(id))
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to check if contextual metric exits", logger.ErrField(err))
			return
		}
		if ie {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgIdentExists))
			return
		}

		cmetric.Id = int64(id)

		exists, err := api.PG.UpdateContextualMetric(ctx, cmetric)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to update contextual metric", logger.ErrField(err))
			return
		}
		if !exists {
			c.JSON(http.StatusNotFound, tools.MsgRes(tools.MsgContextualMetricNotFound))
			return
		}
		api.Log.Info("Contextual metric updated, id: " + rawId)

		c.JSON(http.StatusOK, tools.EmptyRes())
	}
}
