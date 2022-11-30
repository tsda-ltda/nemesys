package service

import (
	"log"

	"github.com/fernandotsda/nemesys/shared/env"
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

func Start(name string, instantiator func() Service, setups ...func(Service)) {
	err := env.LoadEnvFile()
	if err != nil {
		log.Printf("Fail to load enviroment file, err:%s", err)
	}
	env.Init()
	service := instantiator()
	for _, setup := range setups {
		setup(service)
	}
	service.Run()
}
