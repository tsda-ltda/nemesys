package main

import (
	"log"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/router"
	"github.com/joho/godotenv"
)

func main() {
	// load enviroment
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("fail to load enviroment, err: %s", err)
	}

	// create api
	api, err := api.New()
	if err != nil {
		log.Fatalf("fail to create api, err :  %s", err)
	}
	defer api.Close()

	// set routes
	router.Set(api)

	// listen and server
	err = api.Run(":9000")
	log.Fatalf("finished, err: %s", err)
}
