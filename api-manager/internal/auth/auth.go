package auth

import (
	"fmt"
	"strconv"
	"time"

	"github.com/fernandotsda/nemesys/shared/env"
	"github.com/go-redis/redis/v8"
)

type Auth struct {
	// Redis database client.
	rdb *redis.Client

	// Session Time To Live
	SessionTTL time.Duration

	// Session Token size
	TokenSize int
}

type AuthConfig struct {
}

// Creates a new authentication handler
func New(rdb *redis.Client) (*Auth, error) {
	ttl, err := strconv.Atoi(env.UserSessionTTL)
	if err != nil {
		return nil, fmt.Errorf("fail to parse env.UserSessionTTL to int, err:%s", err)
	}
	size, err := strconv.Atoi(env.UserSessionTokenSize)
	if err != nil {
		return nil, fmt.Errorf("fail to parse env.UserSessionTokenSize to int, err:%s", err)
	}
	return &Auth{
		rdb:        rdb,
		SessionTTL: time.Second * time.Duration(ttl),
		TokenSize:  size,
	}, nil
}

func (a *Auth) Close() error {
	return a.rdb.Close()
}
