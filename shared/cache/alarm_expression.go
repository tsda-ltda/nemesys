package cache

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/rdb"
	"github.com/go-redis/redis/v8"
)

type GetMetricAlarmExpressionsResponse struct {
	// Exists is the cache existence.
	Exists bool
	// AlarmExpressions is the alarm expressions.
	AlarmExpressions []models.AlarmExpressionSimplified
}

func (c *Cache) SetMetricAlarmExpressions(ctx context.Context, metricId int64, exp []models.AlarmExpressionSimplified) (err error) {
	b, err := c.encode(exp)
	if err != nil {
		return err
	}
	return c.redis.Set(ctx, rdb.CacheMetricAlarmExpressionsKey(metricId), b, c.metricAlarmExpressionsExp).Err()
}

func (c *Cache) GetMetricAlarmExpressions(ctx context.Context, metricId int64) (r GetMetricAlarmExpressionsResponse, err error) {
	r.AlarmExpressions = []models.AlarmExpressionSimplified{}
	b, err := c.redis.Get(ctx, rdb.CacheMetricAlarmExpressionsKey(metricId)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return r, nil
		}
		return r, err
	}
	r.Exists = true
	return r, c.decode(b, &r.AlarmExpressions)
}

func (c *Cache) SetMetricsAlarmExpressions(ctx context.Context, metricIds []int64, exp [][]models.AlarmExpressionSimplified) (err error) {
	pipe := c.redis.Pipeline()
	for i, id := range metricIds {
		b, err := c.encode(exp[i])
		if err != nil {
			return err
		}
		pipe.Set(ctx, rdb.CacheMetricAlarmExpressionsKey(id), b, c.metricAlarmExpressionsExp)
	}
	_, err = pipe.Exec(ctx)
	return err
}

func (c *Cache) GetMetricsAlarmExpressions(ctx context.Context, metricsIds []int64) (expressions []GetMetricAlarmExpressionsResponse, err error) {
	expressions = make([]GetMetricAlarmExpressionsResponse, len(metricsIds))
	cmds := make([]*redis.StringCmd, len(metricsIds))
	pipe := c.redis.Pipeline()
	for i, id := range metricsIds {
		cmds[i] = pipe.Get(ctx, rdb.CacheMetricAlarmExpressionsKey(id))
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
		err = c.decode(b, &expressions[i].AlarmExpressions)
		if err != nil {
			return nil, err
		}
		expressions[i].Exists = true
	}
	return expressions, nil
}
