package ctxmetric

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/amqph"
	"github.com/fernandotsda/nemesys/shared/influxdb"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/gin-gonic/gin"
)

var (
	ErrFailToGetCustomQueryOnCache    = errors.New("fail to get custom query on cache")
	ErrFailToSetCustomQueryOnCache    = errors.New("fail to set custom query on cache")
	ErrFailToGetCustomQueryOnDatabase = errors.New("fail to get custom query on database")
	ErrCustomQueryNotFound            = errors.New("custom query does not exists")
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
		cq, err := GetCustomQueryFlux(api, c)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			if err == ErrCustomQueryNotFound {
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

// GetCustomQueryFlux get the custom query id/ident on gin context query. Try to get
// the flux on cache, if cache is missing goes to database and save on cache after.
func GetCustomQueryFlux(api *api.API, c *gin.Context) (flux string, err error) {
	ctx := c.Request.Context()
	rawCustomQuery := c.Query("custom_query")
	if len(rawCustomQuery) != 0 {
		id, err := strconv.ParseInt(rawCustomQuery, 0, 32)
		if err != nil {
			cacheRes, err := api.Cache.GetCustomQueryByIdent(ctx, rawCustomQuery)
			if err != nil {
				return flux, ErrFailToGetCustomQueryOnCache
			}
			if !cacheRes.Exists {
				dbRes, err := api.PG.GetCustomQueryByIdent(ctx, rawCustomQuery)
				if err != nil {
					return flux, ErrFailToGetCustomQueryOnDatabase
				}
				if !dbRes.Exists {
					return flux, ErrCustomQueryNotFound
				}
				flux = dbRes.CustomQuery.Flux
				err = api.Cache.SetCustomQueryByIdent(ctx, dbRes.CustomQuery.Flux, rawCustomQuery)
				if err != nil {
					return flux, ErrFailToSetCustomQueryOnCache
				}
			} else {
				flux = cacheRes.Flux
			}
		} else {
			cacheRes, err := api.Cache.GetCustomQuery(ctx, int32(id))
			if err != nil {
				return flux, ErrFailToGetCustomQueryOnCache
			}

			if !cacheRes.Exists {
				dbRes, err := api.PG.GetCustomQuery(ctx, int32(id))
				if err != nil {
					return flux, ErrFailToGetCustomQueryOnDatabase
				}
				if !dbRes.Exists {
					return flux, ErrCustomQueryNotFound
				}
				flux = dbRes.CustomQuery.Flux
				err = api.Cache.SetCustomQuery(ctx, dbRes.CustomQuery.Flux, int32(id))
				if err != nil {
					return flux, ErrFailToSetCustomQueryOnCache
				}
			} else {
				flux = cacheRes.Flux
			}
		}
	}
	return flux, nil
}
