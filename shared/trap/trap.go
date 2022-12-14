package trap

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/fernandotsda/nemesys/shared/amqph"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	g "github.com/gosnmp/gosnmp"
)

type Config struct {
	models.TrapListener
	// Amqph is the amqp handler.
	Amqph *amqph.Amqph
	// Logger is the logger.
	Logger *logger.Logger
	// ServiceIdent is the service ident.
	ServiceIdent string
}

type Trap struct {
	// tl is the trap listener.
	tl models.TrapListener
	// config is the trap config.
	config Config
	// amqph is the amqp handler.
	amqph *amqph.Amqph
	// log is the log.
	log *logger.Logger
	// done is the done channel.
	done chan struct{}
	// listenMu is the listener mutex.
	listenMu sync.Mutex
}

func New(config Config) *Trap {
	trap := &Trap{
		tl:     config.TrapListener,
		config: config,
		log:    config.Logger,
		amqph:  config.Amqph,
		done:   make(chan struct{}),
	}
	go trap.run()
	return trap
}

func (t *Trap) run() {
	params := g.Default
	params.Community = t.tl.Community
	params.Transport = t.tl.Transport

	tl := g.NewTrapListener()
	defer tl.Close()

	tl.OnNewTrap = t.handler
	tl.Params = params

	addr := fmt.Sprintf("%s:%d", t.tl.Host, t.tl.Port)

	go func() {
		t.listenMu.Lock()
		defer t.listenMu.Unlock()

		t.log.Info(fmt.Sprintf("Listening traps at: %s for category id: %d", addr, t.tl.AlarmCategoryId))
		err := tl.Listen(addr)
		if err != nil {
			t.log.Error("Trap listener stopped to listen on addr: "+addr, logger.ErrField(err))
			return
		}
	}()
	<-t.done
}

func (t *Trap) Update(tl models.TrapListener) {
	var restartListener bool
	if t.tl.Host != tl.Host || t.tl.Port != tl.Port {
		restartListener = true
	}
	t.tl = tl
	t.log.Info("Trap listener updated, id: " + strconv.Itoa(int(t.GetId())))

	if restartListener {
		t.log.Info("Restarting trap listener because host, port or transport was updated")
		t.Close()
		go t.run()
	}
}

func (t *Trap) GetId() int32 {
	return t.tl.Id
}

// Close closes the listener, but not the connections passed on the configuration.
func (t *Trap) Close() {
	t.log.Info("Closing trap listener, id: " + strconv.Itoa(int(t.GetId())))
	t.done <- struct{}{}
}
