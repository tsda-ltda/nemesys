package main

import (
	dhs "github.com/fernandotsda/nemesys/data-history"
	"github.com/fernandotsda/nemesys/shared/service"
)

func main() {
	service.Start(func() service.Service {
		return dhs.New()
	})
}
