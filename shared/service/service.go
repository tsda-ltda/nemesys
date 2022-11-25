package service

import (
	"log"
	"time"

	"github.com/fernandotsda/nemesys/shared/env"
	recovery "github.com/fernandotsda/nemesys/shared/service-recovery"
)

type Tools struct {
	doneChs []chan error
}

func NewTools() Tools {
	return Tools{
		doneChs: make([]chan error, 0),
	}
}

func (st *Tools) DispatchDone(err error) {
	for _, v := range st.doneChs {
		v <- err
	}
}

func (st *Tools) Done() <-chan error {
	c := make(chan error)
	st.doneChs = append(st.doneChs, c)
	return c
}

type Service interface {
	Run()
	Done() <-chan error
	DispatchDone(error)
	Close() error
}

func Start(instantiator func() Service, recoveryOptions ...recovery.Options) {
	if len(recoveryOptions) > 1 {
		log.Fatal("Invalid recovery options lenght")
		return
	}

	var opts recovery.Options
	if len(recoveryOptions) == 0 {
		opts = recovery.Options{
			MaxRecovers:          5,
			RecoverTimeout:       time.Second * 30,
			ResetRecoversTimeout: time.Minute * 5,
		}
	} else {
		opts = recoveryOptions[0]
	}

	err := env.LoadEnvFile()
	if err != nil {
		log.Printf("Fail to load enviroment file, err:%s", err)
	}
	env.Init()

	service := instantiator()
	recovery.Run(service, opts)
}
