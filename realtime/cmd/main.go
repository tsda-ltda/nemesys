package main

import (
	"log"
	"time"

	rts "github.com/fernandotsda/nemesys/realtime"
	"github.com/fernandotsda/nemesys/shared/env"
	recovery "github.com/fernandotsda/nemesys/shared/service-recovery"
)

func main() {
	err := env.LoadEnvFile()
	if err != nil {
		log.Println("fail to load env file, err: " + err.Error())
	}
	env.Init()
	recovery.Run(start, recovery.ServiceRecoveryConfig{
		MaxRecovers:          5,
		RecoverTimeout:       time.Second * 5,
		ResetRecoversTimeout: time.Minute,
	})
}

func start() {
	s := rts.New()
	defer s.Close()
	s.Run()
}
