package amqph

import (
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/rabbitmq/amqp091-go"
)

// Amqph is an amqp handler for the api manager packages.
type Amqph struct {
	// conn is the amqp connection.
	conn *amqp091.Connection
	// log is the Logger.
	log *logger.Logger
	// plumber is the plumber for responses.
	plumber *models.AMQPPlumber
	// rtsMetricDataRequestsCh is the channel to request metric data.
	rtsMetricDataRequestsCh chan amqp091.Publishing
	// stopRTSMetricDataListenerCh is the channel to stop the RTS metric data listener.
	stopRTSMetricDataListenerCh chan any
}

// New returns a new Amqph.
func New(conn *amqp091.Connection, log *logger.Logger) *Amqph {
	amqph := &Amqph{
		conn:                        conn,
		log:                         log,
		plumber:                     models.NewAMQPPlumber(),
		rtsMetricDataRequestsCh:     make(chan amqp091.Publishing),
		stopRTSMetricDataListenerCh: make(chan any),
	}

	// run listeners
	go amqph.listenRTSMetricData()
	go amqph.listenRTSMetricDataRequests()
	return amqph
}
