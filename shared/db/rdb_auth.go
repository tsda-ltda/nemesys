package db

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/fernandotsda/nemesys/shared/env"
	"github.com/go-redis/redis/v8"
)

func RDBAuthConnect() (c *redis.Client, err error) {
	ctx := context.Background()

	// get redis auth database number
	db, err := strconv.Atoi(env.RDBAuthDB)
	if err != nil {
		return nil, fmt.Errorf("\nfail to parse to int RDB_DB env, err: %s", err)
	}

	// redis client
	c = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("RDB_AUTH_URL"),
		DB:       db,
		Password: os.Getenv("RDB_AUTH_PW"),
	})

	// ping redis
	if c.Conn(ctx).Ping(ctx).Err() != nil {
		return nil, fmt.Errorf("fail to ping redis, err: %s", err)
	}
	return c, err
}
