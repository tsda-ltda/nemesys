package main

import (
	"context"
	"log"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/router"
	"github.com/fernandotsda/nemesys/api-manager/internal/user"
	"github.com/fernandotsda/nemesys/shared/env"
)

func main() {
	// load enviroment
	env.Init()

	// create api
	api, err := api.New()
	if err != nil {
		log.Fatalf("%s", err)
	}
	defer api.Close()

	// create default user
	err = user.CreateDefaultUser(context.Background(), api)
	if err != nil {
		log.Fatalf("fail to create default user, err: %s", err)
	}

	// set routes
	router.Set(api)

	// listen and server
	err = api.Run()
	log.Fatalf("finished, err: %s", err)
}
