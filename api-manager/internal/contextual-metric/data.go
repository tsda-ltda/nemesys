package ctxmetric

import (
	"net/http"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/amqph"
	"github.com/fernandotsda/nemesys/shared/influxdb"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/gin-gonic/gin"
)

// Retunrn the current metric's value.
// Responses:
//   - 503 If data is not available.
//   - 200 If succeeded.
func DataHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		// get metric request
		r, err := tools.GetMetricRequest(c)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to get metric request", logger.ErrField(err))
			return
		}

		// fetch data
		d, err := api.Amqph.GetRTSData(r, api.GetServiceIdent())
		if err != nil {
			if err == amqph.ErrRequestTimeout {
				c.JSON(http.StatusServiceUnavailable, tools.JSONMSG(tools.MsgRequestTimeout))
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to publish data request", logger.ErrField(err))
			return
		}

		// parse type
		t := amqp.ToMessageType(d.Type)

		// check if something is wrong
		if t != amqp.OK {
			c.JSON(amqp.ParseToHttpStatus(t), tools.JSONMSG(amqp.GetMessage(t)))
			return
		}

		// parse body
		var data models.MetricDataResponse
		err = amqp.Decode(d.Body, &data)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to decode amqp body", logger.ErrField(err))
			return
		}

		c.JSON(http.StatusOK, models.Data{
			Value: data.Value,
		})
	}
}

func QueryDataHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		r, err := tools.GetMetricRequest(c)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to get metric request", logger.ErrField(err))
			return
		}

		var opts influxdb.QueryOptions
		opts.MetricId = r.MetricId
		opts.MetricType = r.MetricType
		opts.DataPolicyId = r.DataPolicyId

		opts.Start = c.Query("start")
		opts.Stop = c.Query("stop")
		cq, err := tools.GetCustomQueryFlux(api, c)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			if err == tools.ErrCustomQueryNotFound {
				c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgCustomQueryNotFound))
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to get custom query", logger.ErrField(err))
			return
		}
		opts.CustomQueryFlux = cq
		points, err := api.Influx.Query(ctx, opts)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			if err == influxdb.ErrInvalidQueryOptions {
				c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to query metric data", logger.ErrField(err))
			return
		}
		c.JSON(http.StatusOK, points)
	}
}
