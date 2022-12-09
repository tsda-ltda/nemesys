package cache

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/rdb"
	"github.com/go-redis/redis/v8"
)

type ServerCostResultResponse struct {
	Exists bool
	Result models.ServerCostResult
}

func (c *Cache) SetServerCostResult(ctx context.Context, result models.ServerCostResult) (err error) {
	b, err := c.encode(result)
	if err != nil {
		return err
	}
	return c.redis.Set(ctx, rdb.CacheServerCostResultKey(), b, c.customQueryExp).Err()
}

func (c *Cache) GetServerCostResult(ctx context.Context) (r ServerCostResultResponse, err error) {
	b, err := c.redis.Get(ctx, rdb.CacheServerCostResultKey()).Bytes()
	if err != nil {
		if err == redis.Nil {
			return r, nil
		}
		return r, err
	}
	return r, c.decode(b, &r.Result)
}
