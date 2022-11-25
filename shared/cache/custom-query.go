package cache

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/rdb"
	"github.com/go-redis/redis/v8"
)

type GetCustomQueryResponse struct {
	Exists bool
	Flux   string
}

func (c *Cache) SetCustomQuery(ctx context.Context, flux string, id int32) error {
	return c.redis.Set(ctx, rdb.CacheCustomQueryKey(id), flux, c.customQueryExp).Err()
}

func (c *Cache) GetCustomQuery(ctx context.Context, id int32) (r GetCustomQueryResponse, err error) {
	cmd := c.redis.Get(ctx, rdb.CacheCustomQueryKey(id))
	err = cmd.Err()
	if err != nil {
		if err == redis.Nil {
			return r, nil
		}
		return r, err
	}
	r.Exists = true
	return r, cmd.Scan(&r.Flux)
}

func (c *Cache) SetCustomQueryByIdent(ctx context.Context, flux string, ident string) error {
	return c.redis.Set(ctx, rdb.CacheCustomQueryByIdentKey(ident), flux, c.customQueryExp).Err()
}

func (c *Cache) GetCustomQueryByIdent(ctx context.Context, ident string) (r GetCustomQueryResponse, err error) {
	cmd := c.redis.Get(ctx, rdb.CacheCustomQueryByIdentKey(ident))
	err = cmd.Err()
	if err != nil {
		if err == redis.Nil {
			return r, nil
		}
		return r, err
	}
	r.Exists = true
	return r, cmd.Scan(&r.Flux)
}
