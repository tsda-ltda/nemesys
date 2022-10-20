package main

import (
	"github.com/fernandotsda/nemesys/shared/env"
	"github.com/fernandotsda/nemesys/translators/snmp"
)

func main() {
	// load enviroment
	env.LoadEnvFile()
	env.Init()

	// run service
	s := snmp.New()
	s.Run()

	defer s.Close()
	<-s.Done
}
