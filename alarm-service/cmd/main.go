package main

import (
	"github.com/fernandotsda/nemesys/alarm-service"
	"github.com/fernandotsda/nemesys/shared/service"
)

func main() {
	service.Start(service.Alarm, alarm.New)
}
