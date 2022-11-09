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
	// containerNotifierCh is the channel to notify a container creation or update.
	containerNotifierCh chan containerNotification
	// metricNotifierCh is the channel to notify a metric creation or update.
	metricNotifierCh chan any
}

// New returns a new Amqph.
func New(conn *amqp091.Connection, log *logger.Logger) *Amqph {
	amqph := &Amqph{
		conn:                    conn,
		log:                     log,
		plumber:                 models.NewAMQPPlumber(),
		rtsMetricDataRequestsCh: make(chan amqp091.Publishing),
		containerNotifierCh:     make(chan containerNotification),
		metricNotifierCh:        make(chan any),
	}

	// run listeners
	go amqph.listenRTSMetricData()
	go amqph.listenRTSMetricDataRequests()
	go amqph.containerNotifier()
	go amqph.metricNotifier()
	return amqph
}
