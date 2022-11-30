package middleware

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/gin-gonic/gin"
)

// ParseContextParams parses, if they are idents, team's params: id and ctxId.
func ParseContextParams(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		teamRawId := c.Param("teamId")
		ctxRawId := c.Param("ctxId")

		_, err1 := strconv.ParseInt(teamRawId, 10, 32)
		_, err2 := strconv.ParseInt(ctxRawId, 10, 32)

		if (err1 != nil) != (err2 != nil) {
			c.AbortWithStatusJSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgParamsNotSameType))
			return
		}
		if err1 == nil {
			c.Next()
			return
		}

		r, err := api.PG.GetContextTreeId(ctx, ctxRawId, teamRawId)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		if !r.Exists {
			c.AbortWithStatusJSON(http.StatusNotFound, tools.JSONMSG(tools.MsgContextNotFound))
			return
		}

		for i, v := range c.Params {
			switch v.Key {
			case "ctxId":
				c.Params[i] = gin.Param{
					Key:   "ctxId",
					Value: strconv.FormatInt(int64(r.ContextId), 10),
				}
			case "id":
				c.Params[i] = gin.Param{
					Key:   "id",
					Value: strconv.FormatInt(int64(r.TeamId), 10),
				}
			}
		}
		c.Next()
	}
}

// ParseContextualMetricParams parses, if they are idents, all team's params: id, ctxId, and metricId.
func ParseContextualMetricParams(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		metricRawId := c.Param("metricId")
		teamRawId := c.Param("teamId")
		ctxRawId := c.Param("ctxId")

		_, err1 := strconv.ParseInt(metricRawId, 10, 64)
		_, err2 := strconv.ParseInt(teamRawId, 10, 32)
		_, err3 := strconv.ParseInt(ctxRawId, 10, 32)

		if (err1 != nil) != (err2 != nil) || (err1 != nil) != (err3 != nil) {
			c.AbortWithStatusJSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgParamsNotSameType))
			return
		}
		if err1 == nil {
			c.Next()
			return
		}

		r, err := api.PG.GetContextualMetricTreeId(ctx, metricRawId, ctxRawId, teamRawId)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.AbortWithStatus(http.StatusInternalServerError)
			api.Log.Error("Fail to get ids on database", logger.ErrField(err))
			return
		}
		if !r.Exists {
			c.AbortWithStatusJSON(http.StatusNotFound, tools.JSONMSG(tools.MsgContextualMetricNotFound))
			return
		}

		for i, v := range c.Params {
			switch v.Key {
			case "id":
				c.Params[i] = gin.Param{
					Key:   "id",
					Value: strconv.FormatInt(int64(r.TeamId), 10),
				}
			case "metricId":
				c.Params[i] = gin.Param{
					Key:   "metricId",
					Value: strconv.FormatInt(r.ContextualMetricId, 10),
				}
			case "ctxId":
				c.Params[i] = gin.Param{
					Key:   "ctxId",
					Value: strconv.FormatInt(int64(r.ContextId), 10),
				}
			}
		}
		c.Next()
	}
}
