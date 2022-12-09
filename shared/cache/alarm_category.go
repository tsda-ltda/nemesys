package cache

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/rdb"
	"github.com/go-redis/redis/v8"
)

func (c *Cache) SetAlarmCategoriesSimplified(ctx context.Context, categories []models.AlarmCategorySimplified) (err error) {
	pipe := c.redis.Pipeline()
	for _, category := range categories {
		b, err := c.encode(category)
		if err != nil {
			return err
		}
		pipe.Set(ctx, rdb.CacheAlarmCategoryKey(category.Id), b, c.metricAlarmCategoryExp)
	}
	_, err = pipe.Exec(ctx)
	return err
}

func (c *Cache) GetAlarmCategoriesSimplified(ctx context.Context, ids []int32) (categories []models.AlarmCategorySimplified, err error) {
	categories = make([]models.AlarmCategorySimplified, 0, len(ids))
	cmds := make([]*redis.StringCmd, len(ids))
	pipe := c.redis.Pipeline()
	for i, id := range ids {
		cmds[i] = pipe.Get(ctx, rdb.CacheAlarmCategoryKey(id))
	}
	_, err = pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, err
	}
	for i, cmd := range cmds {
		b, err := cmd.Bytes()
		if err != nil {
			if err == redis.Nil {
				continue
			}
			return nil, err
		}
		var category models.AlarmCategorySimplified
		err = c.decode(b, &category)
		if err != nil {
			return nil, err
		}
		category.Id = ids[i]
		categories = append(categories, category)

	}
	return categories, nil
}
