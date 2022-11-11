package ctxmetric

import (
	"net/http"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/amqph"
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
