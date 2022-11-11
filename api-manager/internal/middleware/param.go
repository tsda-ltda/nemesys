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

		// get params
		teamRawId := c.Param("id")
		ctxRawId := c.Param("ctxId")

		// parse to number
		_, err1 := strconv.ParseInt(teamRawId, 10, 32)
		_, err2 := strconv.ParseInt(ctxRawId, 10, 32)

		// check if params are differents types
		if (err1 != nil) != (err2 != nil) {
			c.AbortWithStatusJSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgParamsNotSameType))
			return
		}

		// check if is id
		if err1 == nil {
			c.Next()
			return
		}

		// get ids
		r, err := api.PgConn.Contexts.GetIdsByIdent(ctx, ctxRawId, teamRawId)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		// check if exists
		if !r.Exists {
			c.AbortWithStatusJSON(http.StatusNotFound, tools.JSONMSG(tools.MsgContextNotFound))
			return
		}

		// set params
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

		// get params
		metricRawId := c.Param("metricId")
		teamRawId := c.Param("id")
		ctxRawId := c.Param("ctxId")

		// parse to number
		_, err1 := strconv.ParseInt(metricRawId, 10, 64)
		_, err2 := strconv.ParseInt(teamRawId, 10, 32)
		_, err3 := strconv.ParseInt(ctxRawId, 10, 32)

		// check if params are different types
		if (err1 != nil) != (err2 != nil) || (err1 != nil) != (err3 != nil) {
			c.AbortWithStatusJSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgParamsNotSameType))
			return
		}

		// check if could parse id to number
		if err1 == nil {
			c.Next()
			return
		}

		// fetch ids on database
		r, err := api.PgConn.ContextualMetrics.GetIdsByIdent(ctx, metricRawId, ctxRawId, teamRawId)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			api.Log.Error("fail to get ids on database", logger.ErrField(err))
			return
		}

		// check if exists
		if !r.Exists {
			c.AbortWithStatusJSON(http.StatusNotFound, tools.JSONMSG(tools.MsgContextualMetricNotFound))
			return
		}

		// set params
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
