package recovery

import (
	"log"
	"os"
	"time"

	"github.com/fernandotsda/nemesys/shared/env"
)

type ServiceRecoveryConfig struct {
	// MaxRecovers is the maximum number of recoveries before
	// the ResetRecoversTimeout is reached.
	MaxRecovers int
	// RecoverTimeout is the timeout to call the Handler after
	// a recover.
	RecoverTimeout time.Duration
	// ResetRecoversTimeout is the timeout to reset the current
	// numbers of recovers to not reach the maximum value. Any
	// new recovery will reset this timeout.
	ResetRecoversTimeout time.Duration
}

type ServiceRecovery struct {
	// config is the service recovery configuration.
	config ServiceRecoveryConfig
	// recovers is the current number of recovers.
	recovers int
	// resetRecoversTicker is the ticker to reset
	// the current number of recovers. If any new
	// recover resets the ticker.
	resetRecoversTicker *time.Ticker
	// handler is the handler called every recover.
	handler func()
}

func (st *ServiceRecovery) recover() {
	if r := recover(); r != nil {
		log.Println("Service recover system recovered service from a panic, reason: ", r)
		if st.recovers >= st.config.MaxRecovers {
			log.Fatalln("Max recovers reached, exiting...")
			return
		}

		log.Println("Service entering in timeout mode...")
		time.Sleep(st.config.RecoverTimeout)

		st.recovers++
		st.resetRecoversTicker.Reset(st.config.ResetRecoversTimeout)

		log.Printf("Running service again after: %f seconds.", st.config.RecoverTimeout.Seconds())
		st.exec()
	}
}

func (sr *ServiceRecovery) exec() {
	defer sr.recover()
	sr.handler()
}

func (sr *ServiceRecovery) reseter() {
	for range sr.resetRecoversTicker.C {
		sr.recovers = 0
	}
}

func Run(handler func(), config ServiceRecoveryConfig) {
	f, err := os.OpenFile(env.ServiceRecoveryLogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	sr := ServiceRecovery{
		config:              config,
		resetRecoversTicker: time.NewTicker(config.ResetRecoversTimeout),
		handler:             handler,
	}
	go sr.reseter()
	sr.exec()
}
