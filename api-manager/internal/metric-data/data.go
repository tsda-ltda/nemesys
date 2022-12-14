package metricdata

import (
	"net/http"
	"strconv"
	"time"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/amqph"
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
//   - 200 If succeeded.
func AddHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		var data models.MetricDataByRefkey
		err := c.ShouldBind(&data)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}

		err = api.Validate.Struct(data)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidBody))
			return
		}

		var form models.BasicMetricAddDataForm

		cacheRes, err := api.Cache.GetMetricAddDataForm(ctx, data.Refkey)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to get metricAddDataForm on cache", logger.ErrField(err))
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
				api.Log.Error("Fail to get refkey", logger.ErrField(err))
				return
			}
			if !exists {
				c.JSON(http.StatusNotFound, tools.MsgRes(tools.MsgRefkeyNotFound))
				return
			}

			r, err := api.PG.GetMetricRequest(ctx, rk.MetricId)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				c.Status(http.StatusInternalServerError)
				api.Log.Error("Fail to get metric request", logger.ErrField(err))
				return
			}

			_, enabled, err := api.PG.GetMetricDHSEnabled(ctx, rk.MetricId)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				c.Status(http.StatusInternalServerError)
				api.Log.Error("Fail to check if metric dhs_enabled is enabled", logger.ErrField(err))
				return
			}

			form.MetricId = r.MetricRequest.MetricId
			form.DataPolicyId = r.MetricRequest.DataPolicyId
			form.MetricType = r.MetricRequest.MetricType
			form.ContainerId = r.MetricRequest.ContainerId
			form.Enabled = r.Enabled
			form.DHSEnabled = enabled

			err = api.Cache.SetMetricAddDataForm(ctx, data.Refkey, form)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				c.Status(http.StatusInternalServerError)
				api.Log.Error("Fail to set metricAddDataForm on cache", logger.ErrField(err))
				return
			}
		}

		if !form.Enabled {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgMetricDisabled))
			return
		}

		value, err := types.ParseValue(data.Value, form.MetricType)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidMetricData))
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
			var timestamp time.Time
			if data.Timestamp > 0 {
				timestamp = time.Unix(data.Timestamp, 0)
			} else {
				timestamp = time.Now()
			}

			err = api.Influx.WritePoint(ctx, metricDataResponse, timestamp)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				c.Status(http.StatusInternalServerError)
				api.Log.Error("Fail to write point in influxdb", logger.ErrField(err))
				return
			}
			api.Log.Debug("Metric data point saved on influxdb, metric id: " + metricIdString)
		}

		b, err := amqp.Encode(metricDataResponse)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to encode metric data response", logger.ErrField(err))
			return
		}

		api.Amqph.Publish(amqph.Publish{
			Exchange: amqp.ExchangeCheckMetricAlarm,
			Publishing: amqp091.Publishing{
				Body: b,
				Type: amqp.FromMessageType(amqp.OK),
			},
		})

		api.Amqph.Publish(amqph.Publish{
			Exchange:   amqp.ExchangeMetricDataRes,
			RoutingKey: "rts",
			Publishing: amqp091.Publishing{
				Body: b,
				Type: amqp.FromMessageType(amqp.OK),
			},
		})

		api.Log.Debug("Metric data sent to Alarm service and RTS, metric id: " + metricIdString)
	}
}
