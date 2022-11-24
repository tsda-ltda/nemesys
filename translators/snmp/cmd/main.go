package main

import (
	"log"
	"time"

	"github.com/fernandotsda/nemesys/shared/env"
	recovery "github.com/fernandotsda/nemesys/shared/service-recovery"
	"github.com/fernandotsda/nemesys/translators/snmp"
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
	// run service
	s := snmp.New()
	s.Run()
}
