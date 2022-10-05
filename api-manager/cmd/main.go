package main

import (
	"context"
	"log"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/router"
	"github.com/fernandotsda/nemesys/api-manager/internal/user"
	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/env"
	"github.com/fernandotsda/nemesys/shared/logger"
)

func main() {
	// load enviroment file
	err := env.LoadEnvFile()
	if err != nil {
		log.Printf("fail to load env file, err: %s", err)
	}

	// load enviroment
	env.Init()

	// connect to amqp server
	conn, err := amqp.Dial()
	if err != nil {
		log.Fatalf("fail to dial with amqp server, err: %s", err)
	}

	// create logger
	logger, err := logger.New(conn, logger.Config{
		Service:        "api-manager",
		ConsoleLevel:   logger.ParseLevelEnv(env.LogConsoleLevelAPIManager),
		BroadcastLevel: logger.ParseLevelEnv(env.LogBroadcastLevelAPIManager),
	})

	if err != nil {
		log.Fatalf("fail to create logger, err: %s", err)
	}

	// create api
	api, err := api.New(logger)
	if err != nil {
		log.Fatalf("fail to create api, err: %s", err)
	}
	defer api.Close()

	// create default user
	err = user.CreateDefaultUser(context.Background(), api)
	if err != nil {
		logger.Fatal("fail to create default user, err: " + err.Error())
	}
	logger.Info("default master user created with success")

	// set routes
	router.Set(api)

	// listen and server
	err = api.Run()
	logger.Fatal("server finished, err: " + err.Error())
}
