package ctxmetric

import (
	"net/http"
	"strconv"
	"time"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/types"
	"github.com/gin-gonic/gin"
)

// Get the metric's alarm state.
// Responses:
//   - 404 If alarm state not found.
//   - 200 If succeeded.
func GetAlarmStateHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		rawId := c.Param("ctxMetricId")
		id, err := strconv.ParseInt(rawId, 0, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}

		exists, state, err := api.PG.GetAlarmStateByCtxId(ctx, id)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to get alarm state", logger.ErrField(err))
			return
		}
		if !exists {
			c.JSON(http.StatusNotFound, tools.MsgRes(tools.MsgAlarmStateNotFound))
			return
		}

		c.JSON(http.StatusOK, tools.DataRes(state))
	}
}

// Updates the metric's alarm state to recognized.
// Responses:
//   - 404 If alarm state not found.
//   - 400 If alarm state is not in alarm state.
//   - 200 If succeeded.
func RecognizeAlarmStateHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		rawId := c.Param("ctxMetricId")
		id, err := strconv.ParseInt(rawId, 0, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}

		exists, state, err := api.PG.GetAlarmStateByCtxId(ctx, id)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to get alarm state", logger.ErrField(err))
			return
		}
		if !exists {
			c.JSON(http.StatusNotFound, tools.MsgRes(tools.MsgAlarmStateNotFound))
			return
		}

		if state.State != types.ASAlarmed {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgMetricIsNotAlarmed))
			return
		}

		state.LastUpdate = time.Now().Unix()
		state.State = types.ASRecognized

		_, err = api.PG.UpdateAlarmState(ctx, state)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to update alarm state", logger.ErrField(err))
			return
		}
		api.Log.Info("Metric alarm state recognized, ctx metric id: " + rawId)

		c.JSON(http.StatusOK, tools.EmptyRes())
	}
}

// Updates the metric's alarm state to not alarmed.
// Responses:
//   - 404 If alarm state not found.
//   - 400 If alarm state is not in recognized state.
//   - 200 If succeeded.
func ResolveAlarmStateHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		rawId := c.Param("ctxMetricId")
		id, err := strconv.ParseInt(rawId, 0, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}

		exists, state, err := api.PG.GetAlarmStateByCtxId(ctx, id)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to get alarm state", logger.ErrField(err))
			return
		}
		if !exists {
			c.JSON(http.StatusNotFound, tools.MsgRes(tools.MsgAlarmStateNotFound))
			return
		}

		if state.State != types.ASRecognized {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgMetricIsNotRecognized))
			return
		}

		state.LastUpdate = time.Now().Unix()
		state.State = types.ASNotAlarmed

		_, err = api.PG.UpdateAlarmState(ctx, state)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to update alarm state", logger.ErrField(err))
			return
		}
		api.Log.Info("Metric alarm state resolved, ctx metric id: " + rawId)

		c.JSON(http.StatusOK, tools.EmptyRes())
	}
}
