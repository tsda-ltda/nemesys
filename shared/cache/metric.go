package cache

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/rdb"
	"github.com/go-redis/redis/v8"
)

// GetMetricRequestByIdent returns a metric request information.
func (c *Cache) GetMetricRequestByIdent(ctx context.Context, teamIdent string, contextIdent string, metricIdent string) (e bool, r models.MetricRequest, err error) {
	bytes, err := c.redis.Get(ctx, rdb.RDBCacheMetricIdContainerIdKey(teamIdent, contextIdent, metricIdent)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return false, r, nil
		}
		return false, r, err
	}
	err = c.decode(bytes, &r)
	return true, r, err
}

// SetMetricRequestByIdent save a metric request information.
func (c *Cache) SetMetricRequestByIdent(ctx context.Context, teamIdent string, contextIdent string, metricIdent string, ids models.MetricRequest) (err error) {
	bytes, err := c.encode(ids)
	if err != nil {
		return err
	}
	return c.redis.Set(ctx, rdb.RDBCacheMetricIdContainerIdKey(teamIdent, contextIdent, metricIdent), bytes, c.metricIdContainerIdExp).Err()
}

// SetMetricEvExpression save a metric evaluate expression.
func (c *Cache) SetMetricEvExpression(ctx context.Context, metricId int64, expression string) (err error) {
	return c.redis.Set(ctx, rdb.RDBCacheMetricEvExpressionKey(metricId), expression, c.metricEvExpressionExp).Err()
}

// GetMetricEvExpression returns a metric evaluate expression.
func (c *Cache) GetMetricEvExpression(ctx context.Context, metricId int64) (e bool, expression string, err error) {
	expression, err = c.redis.Get(ctx, rdb.RDBCacheMetricEvExpressionKey(metricId)).Result()
	if err == redis.Nil {
		err = nil
	} else if err == nil {
		e = true
	}
	return e, expression, err
}
