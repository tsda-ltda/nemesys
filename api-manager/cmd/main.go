package main

import (
	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/router"
	"github.com/fernandotsda/nemesys/shared/service"
)

func main() {
	service.Start("api-manager", api.New, router.Set)
}
