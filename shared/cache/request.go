package cache

import (
	"context"
	"time"

	"github.com/fernandotsda/nemesys/shared/rdb"
	"github.com/go-redis/redis/v8"
)

func (c *Cache) GetUserLimited(ctx context.Context, ip string) (suspended bool, err error) {
	err = c.redis.Get(ctx, rdb.CacheUserLimited(ip)).Err()
	if err == redis.Nil {
		return false, nil
	}
	if err == nil {
		return true, nil
	}
	return false, err
}

func (c *Cache) SetUserLimited(ctx context.Context, addr string, duration time.Duration) (err error) {
	return c.redis.Set(ctx, rdb.CacheUserLimited(addr), nil, duration).Err()
}
