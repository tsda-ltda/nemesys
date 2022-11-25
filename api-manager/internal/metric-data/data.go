package metricdata

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/types"
	"github.com/gin-gonic/gin"
	"github.com/rabbitmq/amqp091-go"
)

// Adds a metric data to metric data history if enabled and send data
// to real time service.
// Responses:
//   - 400 If invalid body.
//   - 400 If json fields are invalid.
//   - 404 If refkey not found.
//   - 400 If metric is disabled.
//   - 204 If succeeded.
func AddHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		var data models.MetricDataByRefkey
		err := c.ShouldBind(&data)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		err = api.Validate.Struct(data)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidBody))
			return
		}

		var form models.BasicMetricAddDataForm

		cacheRes, err := api.Cache.GetMetricAddDataForm(ctx, data.Refkey)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to get metricAddDataForm on cache", logger.ErrField(err))
			return
		}
		form = cacheRes.Form

		if !cacheRes.Exists {
			exists, rk, err := api.PG.GetRefkey(ctx, data.Refkey)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				c.Status(http.StatusInternalServerError)
				api.Log.Error("fail to get refkey", logger.ErrField(err))
				return
			}
			if !exists {
				c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgRefkeyNotFound))
				return
			}

			r1, err := api.PG.GetMetricRequest(ctx, rk.MetricId)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				c.Status(http.StatusInternalServerError)
				api.Log.Error("fail to get metric request", logger.ErrField(err))
				return
			}

			r2, err := api.PG.GetMetricDHSEnabled(ctx, rk.MetricId)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				c.Status(http.StatusInternalServerError)
				api.Log.Error("fail to check if metric dhs_enabled is enabled", logger.ErrField(err))
				return
			}

			form.MetricId = r1.MetricRequest.MetricId
			form.DataPolicyId = r1.MetricRequest.DataPolicyId
			form.MetricType = r1.MetricRequest.MetricType
			form.ContainerId = r1.MetricRequest.ContainerId
			form.Enabled = r1.Enabled
			form.DHSEnabled = r2.Enabled

			err = api.Cache.SetMetricAddDataForm(ctx, data.Refkey, form)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				c.Status(http.StatusInternalServerError)
				api.Log.Error("fail to set metricAddDataForm on cache", logger.ErrField(err))
				return
			}
		}

		if !form.Enabled {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgMetricDisabled))
			return
		}

		value, err := types.ParseValue(data.Value, form.MetricType)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidMetricData))
			return
		}

		metricDataResponse := models.MetricDataResponse{
			MetricBasicDataReponse: models.MetricBasicDataReponse{
				Id:           form.MetricId,
				Type:         form.MetricType,
				Value:        value,
				DataPolicyId: form.DataPolicyId,
				Failed:       false,
			},
			ContainerId: form.ContainerId,
		}

		metricIdString := strconv.FormatInt(form.MetricId, 10)
		if form.DHSEnabled {
			err = api.Influx.WritePoint(ctx, metricDataResponse)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				c.Status(http.StatusInternalServerError)
				api.Log.Error("fail to write point in influxdb", logger.ErrField(err))
				return
			}
			api.Log.Debug("metric data point saved on influxdb, metric id: " + metricIdString)
		}

		b, err := amqp.Encode(metricDataResponse)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to encode metric data response", logger.ErrField(err))
			return
		}

		api.Amqph.PublisherCh <- models.DetailedPublishing{
			Exchange:   amqp.ExchangeMetricDataResponse,
			RoutingKey: "rts",
			Publishing: amqp091.Publishing{
				Expiration: amqp.DefaultExp,
				Body:       b,
				Type:       amqp.FromMessageType(amqp.OK),
			},
		}
		api.Log.Debug("metric data point sent to RTS, metric id: " + metricIdString)

	}
}
