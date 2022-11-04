package cache

import (
	"context"
	"time"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/db"
	"github.com/go-redis/redis/v8"
)

type Cache struct {
	// redis is the redis client used to get and set data.
	redis *redis.Client

	// metricIdContainerIdExp is the time expire the data.
	metricIdContainerIdExp time.Duration
	// metricEvExpressionExp is the time expire the evaluate expression.
	metricEvExpressionExp time.Duration
}

// New returns a prepared Cache struct.
func New() *Cache {
	// connect to redis
	c, err := db.RDBCacheConnect()
	if err != nil {
		panic("fail to connect to redis cache database")
	}

	return &Cache{
		redis:                  c,
		metricIdContainerIdExp: time.Minute,
		metricEvExpressionExp:  time.Minute,
	}
}

func (c *Cache) Close() {
	c.redis.Close()
}

func (c *Cache) encode(v any) ([]byte, error) {
	return amqp.Encode(v)
}

func (c *Cache) decode(b []byte, v any) error {
	return amqp.Decode(b, v)
}

func (c *Cache) Set(ctx context.Context, b []byte, key string, exp time.Duration) error {
	return c.redis.Set(ctx, key, b, exp).Err()
}

func (c *Cache) Get(ctx context.Context, key string) (b []byte, err error) {
	return c.redis.Get(ctx, key).Bytes()
}