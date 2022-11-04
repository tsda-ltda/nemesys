package main

import (
	"log"

	rts "github.com/fernandotsda/nemesys/realtime"
	"github.com/fernandotsda/nemesys/shared/env"
)

func main() {
	// load enviroment
	err := env.LoadEnvFile()
	if err != nil {
		log.Println("fail to load enviroment file")
	}

	// initialize variables
	env.Init()

	// start server
	s := rts.New()
	defer s.Close()
	s.Run()
}
