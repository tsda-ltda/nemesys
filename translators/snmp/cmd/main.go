package main

import (
	"github.com/fernandotsda/nemesys/shared/service"
	"github.com/fernandotsda/nemesys/translators/snmp"
)

func main() {
	service.Start(func() service.Service {
		return snmp.New()
	})
}
