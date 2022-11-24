package main

import (
	"log"
	"time"

	dhs "github.com/fernandotsda/nemesys/data-history"
	"github.com/fernandotsda/nemesys/shared/env"
	recovery "github.com/fernandotsda/nemesys/shared/service-recovery"
)

func main() {
	// load enviroment
	err := env.LoadEnvFile()
	if err != nil {
		log.Println("fail to load env file")
	}
	env.Init()
	recovery.Run(start, recovery.ServiceRecoveryConfig{
		MaxRecovers:          5,
		RecoverTimeout:       time.Second * 5,
		ResetRecoversTimeout: time.Minute,
	})
}

func start() {
	service, err := dhs.New()
	if err != nil {
		panic(err)
	}
	service.Run()
}
