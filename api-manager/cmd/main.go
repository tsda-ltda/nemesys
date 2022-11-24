package main

import (
	"context"
	"log"
	"time"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/router"
	"github.com/fernandotsda/nemesys/api-manager/internal/user"
	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/env"
	"github.com/fernandotsda/nemesys/shared/initdb"
	_logger "github.com/fernandotsda/nemesys/shared/logger"
	recovery "github.com/fernandotsda/nemesys/shared/service-recovery"
)

func main() {
	err := env.LoadEnvFile()
	if err != nil {
		log.Printf("fail to load env file, err: %s", err)
	}

	env.Init()
	recovery.Run(start, recovery.ServiceRecoveryConfig{
		MaxRecovers:          5,
		RecoverTimeout:       time.Second * 5,
		ResetRecoversTimeout: time.Minute,
	})
}

func start() {
	conn, err := amqp.Dial()
	if err != nil {
		log.Fatalf("fail to dial with amqp server, err: %s", err)
	}

	logger, err := _logger.New(conn, _logger.Config{
		Service:        "api-manager",
		ConsoleLevel:   _logger.ParseLevelEnv(env.LogConsoleLevelAPIManager),
		BroadcastLevel: _logger.ParseLevelEnv(env.LogBroadcastLevelAPIManager),
	})
	if err != nil {
		log.Fatalf("fail to create internal logger, err: %s", err)
	}

	init, err := initdb.PG()
	if err != nil {
		logger.Fatal("fail to initialize database", _logger.ErrField(err))
	}

	api, err := api.New(conn, logger)
	if err != nil {
		logger.Fatal("fail to create api", _logger.ErrField(err))
	}
	defer api.Close()

	if init {
		logger.Info("database initialized")

		err = user.CreateDefaultUser(context.Background(), api)
		if err != nil {
			logger.Fatal("fail to create default user", _logger.ErrField(err))
		}
		logger.Info("default master user created")
	}
	router.Set(api)

	err = api.Run()
	logger.Fatal("server finished, err: " + err.Error())
}
