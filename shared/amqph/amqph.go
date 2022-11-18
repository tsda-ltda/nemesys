package amqph

import (
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/rabbitmq/amqp091-go"
)

// Amqph is an amqp handler common tasks between services.
type Amqph struct {
	// conn is the amqp connection.
	conn *amqp091.Connection
	// log is the Logger.
	log *logger.Logger
	// plumber is the plumber for responses.
	plumber *models.AMQPPlumber
	// PublisherCh is the channel to pubish messages.
	PublisherCh chan models.DetailedPublishing
}

// New returns a new Amqph.
func New(conn *amqp091.Connection, log *logger.Logger) *Amqph {
	amqph := &Amqph{
		conn:        conn,
		log:         log,
		plumber:     models.NewAMQPPlumber(),
		PublisherCh: make(chan models.DetailedPublishing),
	}
	amqph.declareExchages()
	go amqph.publisher()
	return amqph
}

func (a *Amqph) Close() {
	if !a.conn.IsClosed() {
		a.conn.Close()
	}
}
