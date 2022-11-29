package main

import (
	rts "github.com/fernandotsda/nemesys/realtime"
	"github.com/fernandotsda/nemesys/shared/service"
)

func main() {
	service.Start("rts", rts.New)
}
