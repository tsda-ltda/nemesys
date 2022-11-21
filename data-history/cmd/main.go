package main

import (
	"log"

	dhs "github.com/fernandotsda/nemesys/data-history"
	"github.com/fernandotsda/nemesys/shared/env"
)

func main() {
	// load enviroment
	err := env.LoadEnvFile()
	if err != nil {
		log.Println("fail to load env file")
	}
	env.Init()

	service, err := dhs.New()
	if err != nil {
		panic(err)
	}
	service.Run()
}
