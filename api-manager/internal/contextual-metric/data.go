package ctxmetric

import (
	"net/http"
	"strconv"
	"time"

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
			api.Log.Error("fail to get metric request", logger.ErrField(err))
			return
		}

		// fetch data
		d, err := api.Amqph.GetRTSData(r)
		if err != nil {
			if err == amqph.ErrRequestTimeout {
				c.JSON(http.StatusServiceUnavailable, tools.JSONMSG(tools.MsgRequestTimeout))
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to publish data request", logger.ErrField(err))
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
			api.Log.Error("fail to decode amqp body", logger.ErrField(err))
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

		// get metric request
		r, err := tools.GetMetricRequest(c)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to get metric request", logger.ErrField(err))
			return
		}

		now := time.Now().Unix()

		// query options
		var opts influxdb.QueryOptions
		opts.MetricId = r.MetricId
		opts.MetricType = r.MetricType
		opts.DataPolicyId = r.DataPolicyId

		// get start param
		start, err := strconv.ParseInt(c.Query("start"), 0, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}
		opts.Start = influxdb.DurationFromSeconds(start - now)

		// get stop param
		rawStop := c.Param("stop")
		if len(rawStop) > 0 {
			stop, err := strconv.ParseInt(rawStop, 0, 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
				return
			}
			opts.Stop = influxdb.DurationFromSeconds(stop - now)
		} else {
			opts.Stop = "now()"
		}

		// get custom query
		rawCustomQuery := c.Query("custom_query")
		if len(rawCustomQuery) != 0 {
			id, err := strconv.ParseInt(rawCustomQuery, 0, 32)
			if err != nil {
				// get custom query by ident on cache
				cacheRes, err := api.Cache.GetCustomQueryByIdent(ctx, rawCustomQuery)
				if err != nil {
					c.Status(http.StatusInternalServerError)
					api.Log.Error("fail to get custom query on cache", logger.ErrField(err))
					return
				}

				// check if cacke exists
				if !cacheRes.Exists {
					// get custom query by ident on database
					dbRes, err := api.PgConn.CustomQueries.GetByIdent(ctx, rawCustomQuery)
					if err != nil {
						c.Status(http.StatusInternalServerError)
						api.Log.Error("fail to get custom query on cache", logger.ErrField(err))
						return
					}

					// check if custom query exists
					if !dbRes.Exists {
						c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgCustomQueryNotFound))
						return
					}
					opts.CustomQueryFlux = dbRes.CustomQuery.Flux

					// save on cache
					err = api.Cache.SetCustomQueryByIdent(ctx, dbRes.CustomQuery.Flux, rawCustomQuery)
					if err != nil {
						c.Status(http.StatusInternalServerError)
						api.Log.Error("fail to save custom query flux on cache", logger.ErrField(err))
						return
					}
				} else {
					opts.CustomQueryFlux = cacheRes.Flux
				}
			} else {
				// get custom query by id on cache
				cacheRes, err := api.Cache.GetCustomQuery(ctx, int32(id))
				if err != nil {
					c.Status(http.StatusInternalServerError)
					api.Log.Error("fail to get custom query on cache", logger.ErrField(err))
					return
				}

				// check if cacke exists
				if !cacheRes.Exists {
					// get custom query by id on database
					dbRes, err := api.PgConn.CustomQueries.Get(ctx, int32(id))
					if err != nil {
						c.Status(http.StatusInternalServerError)
						api.Log.Error("fail to get custom query on cache", logger.ErrField(err))
						return
					}

					// check if custom query exists
					if !dbRes.Exists {
						c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgCustomQueryNotFound))
						return
					}
					opts.CustomQueryFlux = dbRes.CustomQuery.Flux
					err = api.Cache.SetCustomQuery(ctx, dbRes.CustomQuery.Flux, int32(id))
					if err != nil {
						c.Status(http.StatusInternalServerError)
						api.Log.Error("fail to save custom query flux on cache", logger.ErrField(err))
						return
					}
				} else {
					opts.CustomQueryFlux = cacheRes.Flux
				}
			}
		}

		d, err := api.Influx.Query(ctx, opts)
		if err != nil {
			if err == influxdb.ErrInvalidQueryOptions {
				c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to query metric data", logger.ErrField(err))
			return
		}
		c.JSON(http.StatusOK, d)
	}
}
