package amqph

import (
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/pubsub"
	"github.com/rabbitmq/amqp091-go"
)

const (
	ContainerCreated string = "container_created"
	ContainerUpdated string = "container_updated"
	ContainerDeleted string = "container_deleted"
	MetricCreated    string = "metric_created"
	MetricUpdated    string = "metric_updated"
	MetricDeleted    string = "metrfc_deleted"
)

// Amqph is an amqp handler common tasks between services.
type Amqph struct {
	// Notifications is the a pub/sub of notifications.
	Notifications *pubsub.PubSub
	// conn is the amqp connection.
	conn *amqp091.Connection
	// log is the Logger.
	log *logger.Logger
	// plumber is the plumber for responses.
	plumber *models.AMQPPlumber
	// rtsMetricDataRequestsCh is the channel to request metric data.
	rtsMetricDataRequestsCh chan amqp091.Publishing
	// containerUpdates is the channel to notify a container update.
	containerUpdatesNotifierCh chan ContainerNotification
	// metricUpdatedNotifierCh is the channel to notify a metric update.
	metricUpdatedNotifierCh chan MetricNotification
	// containerCreatedNotifierCh is the channel to notify a container creation.
	containerCreatedNotifierCh chan ContainerNotification
	// metricCreatedNotifierCh is the channel to notify a metric creation.
	metricCreatedNotifierCh chan MetricNotification
	// containerDeletedNotifierCh is the channel to notify that a container was deleted.
	containerDeletedNotifierCh chan int32
	// metricDeletedNotifierCh is the channel to notify that a metric was deleted.
	metricDeletedNotifierCh chan models.MetricPairId
}

// New returns a new Amqph.
func New(conn *amqp091.Connection, log *logger.Logger) *Amqph {
	amqph := &Amqph{
		Notifications:              pubsub.New(),
		conn:                       conn,
		log:                        log,
		plumber:                    models.NewAMQPPlumber(),
		rtsMetricDataRequestsCh:    make(chan amqp091.Publishing),
		containerUpdatesNotifierCh: make(chan ContainerNotification),
		containerCreatedNotifierCh: make(chan ContainerNotification),
		containerDeletedNotifierCh: make(chan int32),
		metricUpdatedNotifierCh:    make(chan MetricNotification),
		metricCreatedNotifierCh:    make(chan MetricNotification),
		metricDeletedNotifierCh:    make(chan models.MetricPairId),
	}

	return amqph
}

func (a *Amqph) Close() {
	if !a.conn.IsClosed() {
		a.conn.Close()
	}
}
