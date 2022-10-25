package db

import (
	"context"
	"fmt"
	"strconv"

	"github.com/fernandotsda/nemesys/shared/env"
	"github.com/go-redis/redis/v8"
)

// Connects to Redis auth database.
func RDBAuthConnect() (c *redis.Client, err error) {
	ctx := context.Background()

	// get redis auth database number
	db, err := strconv.Atoi(env.RDBAuthDB)
	if err != nil {
		return nil, fmt.Errorf("\nfail to parse to int RDB_AUTH_DB env, err: %s", err)
	}

	// redis client
	c = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", env.RDBAuthHost, env.RDBAuthPort),
		DB:       db,
		Password: env.RDBAuthPW,
	})

	// ping redis
	err = c.Conn(ctx).Ping(ctx).Err()
	if err != nil {
		return nil, fmt.Errorf("fail to ping redis, err: %s", err)
	}
	return c, err
}

// Connects to Redis real time service database.
func RDBRTSConnect() (c *redis.Client, err error) {
	ctx := context.Background()

	// get redis rts database number
	db, err := strconv.Atoi(env.RDBRTSDB)
	if err != nil {
		return nil, fmt.Errorf("\nfail to parse to int RDB_RTS_DB env, err: %s", err)
	}

	// redis client
	c = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", env.RDBRTSHost, env.RDBRTSPort),
		DB:       db,
		Password: env.RDBRTSPW,
	})

	// ping redis
	if c.Conn(ctx).Ping(ctx).Err() != nil {
		return nil, fmt.Errorf("fail to ping redis, err: %s", err)
	}
	return c, err
}
