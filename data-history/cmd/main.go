package main

import (
	dhs "github.com/fernandotsda/nemesys/data-history"
	"github.com/fernandotsda/nemesys/shared/service"
)

func main() {
	service.Start(service.DHS, dhs.New)
}
