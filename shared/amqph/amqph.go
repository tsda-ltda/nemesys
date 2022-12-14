package amqph

import (
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/rabbitmq/amqp091-go"
)

type Config struct {
	// Log is the logger.
	Log *logger.Logger
	// Conn is the amqp connection.
	Conn *amqp091.Connection
	// Publishers is the number of publishers.
	Publishers int
}

type Amqph struct {
	// config is the config
	config Config
	// conn is the amqp connection.
	conn *amqp091.Connection
	// log is the Logger.
	log *logger.Logger
	// plumber is the plumber for responses.
	plumber *models.AMQPPlumber
	// doneChs are the done channels.
	doneChs []chan struct{}
	// publisherCh is the channel for publishing.
	publisherCh chan Publish
}

func New(config Config) *Amqph {
	amqph := &Amqph{
		config:      config,
		conn:        config.Conn,
		log:         config.Log,
		plumber:     models.NewAMQPPlumber(),
		doneChs:     []chan struct{}{},
		publisherCh: make(chan Publish),
	}
	amqph.declareExchages()

	for i := 0; i < config.Publishers; i++ {
		go amqph.publisher()
	}

	return amqph
}

func (a *Amqph) done() <-chan struct{} {
	ch := make(chan struct{})
	a.doneChs = append(a.doneChs, ch)
	return ch
}

// Close closes publishers and listeners channels, but not
// the amqp connection or the log.
func (a *Amqph) Close() {
	for _, v := range a.doneChs {
		v <- struct{}{}
		close(v)
	}
}
