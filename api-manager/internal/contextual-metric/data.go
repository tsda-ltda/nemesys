package ctxmetric

import (
	"errors"
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
	"github.com/fernandotsda/nemesys/shared/uuid"
	"github.com/gin-gonic/gin"
	"github.com/rabbitmq/amqp091-go"
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
	p := models.NewAMQPPlumber()
	go func() {
		var options amqph.ListenerOptions
		options.QueueDeclarationOptions.Exclusive = true
		options.QueueBindOptions.Exchange = amqp.ExchangeMetricDataRes
		options.QueueBindOptions.RoutingKey = api.GetServiceIdent()

		msgs, done := api.Amqph.Listen(options)
		for {
			select {
			case d := <-msgs:
				p.Send(d)
			case <-done:
				return
			}
		}
	}()

	return func(c *gin.Context) {
		r, err := tools.GetMetricRequest(c)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to get metric request", logger.ErrField(err))
			return
		}

		b, err := amqp.Encode(r)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to encode amqp body", logger.ErrField(err))
			return
		}

		uuid, err := uuid.New()
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to get new uuid", logger.ErrField(err))
			return
		}

		api.Amqph.Publish(amqph.Publish{
			Exchange:   amqp.ExchangeMetricDataReq,
			RoutingKey: "rts",
			Publishing: amqp091.Publishing{
				Body:          b,
				CorrelationId: uuid,
				Headers:       amqp.RouteHeader(api.GetServiceIdent()),
			},
		})

		d, err := p.Listen(uuid, time.Second*30)
		if err == amqph.ErrRequestTimeout {
			c.JSON(http.StatusServiceUnavailable, tools.MsgRes(tools.MsgRequestTimeout))
			return
		}

		t := amqp.ToMessageType(d.Type)

		if t != amqp.OK {
			c.JSON(amqp.ParseToHttpStatus(t), tools.MsgRes(amqp.GetMessage(t)))
			return
		}

		var data models.MetricDataResponse
		err = amqp.Decode(d.Body, &data)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to decode amqp body", logger.ErrField(err))
			return
		}

		c.JSON(http.StatusOK, tools.DataRes(data.Value))
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

		start, err := strconv.ParseInt(c.Query("start"), 0, 64)
		if err != nil || start < 1 {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}
		opts.Start = start

		stopS := c.Query("stop")
		if stopS != "" {
			stop, err := strconv.ParseInt(stopS, 0, 64)
			if err != nil || stop < 1 {
				c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
				return
			}
			opts.Stop = stop
		}

		cq, err := GetCustomQueryFlux(api, c)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			if err == ErrCustomQueryNotFound {
				c.JSON(http.StatusNotFound, tools.MsgRes(tools.MsgCustomQueryNotFound))
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
				c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to query metric data", logger.ErrField(err))
			return
		}
		c.JSON(http.StatusOK, tools.DataRes(points))
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
				return flux, err
			}
			if !cacheRes.Exists {
				exists, cq, err := api.PG.GetCustomQueryByIdent(ctx, rawCustomQuery)
				if err != nil {
					return flux, err
				}
				if !exists {
					return flux, err
				}
				flux = cq.Flux
				err = api.Cache.SetCustomQueryByIdent(ctx, cq.Flux, rawCustomQuery)
				if err != nil {
					return flux, err
				}
			} else {
				flux = cacheRes.Flux
			}
		} else {
			cacheRes, err := api.Cache.GetCustomQuery(ctx, int32(id))
			if err != nil {
				return flux, err
			}

			if !cacheRes.Exists {
				exists, cq, err := api.PG.GetCustomQuery(ctx, int32(id))
				if err != nil {
					return flux, err
				}
				if !exists {
					return flux, err
				}
				flux = cq.Flux
				err = api.Cache.SetCustomQuery(ctx, cq.Flux, int32(id))
				if err != nil {
					return flux, err
				}
			} else {
				flux = cacheRes.Flux
			}
		}
	}
	return flux, nil
}
