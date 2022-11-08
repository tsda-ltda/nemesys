package ctxmetric

import (
	"net/http"
	"strconv"
	"time"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/uuid"
	"github.com/gin-gonic/gin"
	"github.com/rabbitmq/amqp091-go"
)

// Retunrn the current metric's value.
// Responses:
//   - 503 If data is not available.
//   - 200 If succeeded.
func DataHandler(api *api.API) func(c *gin.Context) {
	// open socket channel
	ch, err := api.Amqp.Channel()
	if err != nil {
		api.Log.Panic("fail to open socket channel", logger.ErrField(err))
		return func(c *gin.Context) {}
	}

	// declare get data exchange
	err = ch.ExchangeDeclare(
		amqp.ExchangeRTSGetMetricData, // name
		"direct",                      // type
		true,                          // durable
		false,                         // auto-deleted
		false,                         // internal
		false,                         // no-wait
		nil,                           // arguments
	)
	if err != nil {
		api.Log.Panic("fail to declare exchange", logger.ErrField(err))
		return func(c *gin.Context) {}
	}

	// declare data exchange
	err = ch.ExchangeDeclare(
		amqp.ExchangeRTSMetricData, // name
		"fanout",                   // type
		true,                       // durable
		false,                      // auto-deleted
		false,                      // internal
		false,                      // no-wait
		nil,                        // arguments
	)
	if err != nil {
		api.Log.Panic("fail to declare exchange", logger.ErrField(err))
		return func(c *gin.Context) {}
	}

	// listen to new data
	go dataListener(api, ch)

	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// get metric id
		id, err := strconv.ParseInt(c.Param("metricId"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		// get metric request
		e, r, err := api.PgConn.ContextualMetrics.GetMetricRequestById(ctx, id)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to get contextual metric, team and context id on database", logger.ErrField(err))
			return
		}

		// check if exists
		if !e {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgContextualMetricNotFound))
			return
		}

		// encode request
		bytes, err := amqp.Encode(r)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to encode amqp body", logger.ErrField(err))
			return
		}

		// generate new unique id
		uuid, err := uuid.New()
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to generate new uuid", logger.ErrField(err))
			return
		}

		// request data
		err = ch.PublishWithContext(ctx,
			amqp.ExchangeRTSGetMetricData, // exchange
			"",                            // routing key
			false,                         // mandatory
			false,                         // immediate
			amqp091.Publishing{
				Expiration:    "30000",
				Body:          bytes,
				CorrelationId: uuid,
			},
		)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to publish data request", logger.ErrField(err))
			return
		}

		// listen to response
		d, err := api.RTSDataPlumber.Listen(uuid, time.Second*30)
		if err != nil {
			c.JSON(http.StatusRequestTimeout, tools.JSONMSG(tools.MsgRequestTimeout))
			api.Log.Warn("RTS plumber data timeout")
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

func dataListener(api *api.API, ch *amqp091.Channel) {
	// declare queue
	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable'
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		api.Log.Panic("fail to declare queue", logger.ErrField(err))
		return
	}

	// bind queue
	err = ch.QueueBind(
		q.Name,                     // queue name
		"",                         // routing key
		amqp.ExchangeRTSMetricData, // exchange
		false,                      // no-wait
		nil,                        // args
	)
	if err != nil {
		api.Log.Panic("fail to bind queue", logger.ErrField(err))
		return
	}

	// consume messages
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		api.Log.Panic("fail to consume messages", logger.ErrField(err))
	}

	// close and cancel channels
	closedCh := make(chan *amqp091.Error)
	canceledCh := make(chan string)
	ch.NotifyCancel(canceledCh)
	ch.NotifyClose(closedCh)

	for {
		select {
		case d := <-msgs:
			api.RTSDataPlumber.Send(d)
		case err := <-closedCh:
			api.Log.Warn("RTS data channel closed", logger.ErrField(err))
			return
		case r := <-canceledCh:
			api.Log.Warn("RTS data channel canceled, reason: " + r)
			return
		}
	}
}
