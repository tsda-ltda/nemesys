package cache

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/rdb"
	"github.com/go-redis/redis/v8"
)

type GetMetricRequestByIdentResponse struct {
	// Exists is the cache existence.
	Exists bool
	// Request is the metric request.
	Request models.MetricRequest
}

type GetMetricRequestResponse struct {
	// Exists is the cache existence.
	Exists bool
	// Request is the metric request.
	Request models.MetricRequest
}

type GetMetricEvExpressionResponse struct {
	// Exists is the cache existence.
	Exists bool
	// Expression is the metric evaluate expression.
	Expression string
}

type GetMetricDataPolicyIdResponse struct {
	// Exists is the cache existence.
	Exists bool
	// DataPolicyId is the data policy id
	DataPolicyId int16
}

// GetMetricRequestByIdent returns a metric request information.
func (c *Cache) GetMetricRequestByIdent(ctx context.Context, teamIdent string, contextIdent string, metricIdent string) (r GetMetricRequestByIdentResponse, err error) {
	bytes, err := c.redis.Get(ctx, rdb.CacheMetricRequestByIdent(teamIdent, contextIdent, metricIdent)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return r, nil
		}
		return r, err
	}
	err = c.decode(bytes, &r.Request)
	r.Exists = true
	return r, err
}

// SetMetricRequestByIdent save a metric request information.
func (c *Cache) SetMetricRequestByIdent(ctx context.Context, teamIdent string, contextIdent string, metricIdent string, ids models.MetricRequest) (err error) {
	bytes, err := c.encode(ids)
	if err != nil {
		return err
	}
	return c.redis.Set(ctx, rdb.CacheMetricRequestByIdent(teamIdent, contextIdent, metricIdent), bytes, c.metricRequestByIdentExp).Err()
}

func (c *Cache) GetMetricRequest(ctx context.Context, id int64) (r GetMetricRequestResponse, err error) {
	b, err := c.redis.Get(ctx, rdb.CacheMetricRequest(id)).Bytes()
	if err != nil {
		if err != redis.Nil {
			return r, err
		}
		return r, nil
	}
	r.Exists = true
	return r, c.decode(b, &r.Request)
}

func (c *Cache) SetMetricRequest(ctx context.Context, request models.MetricRequest) (err error) {
	b, err := c.encode(request)
	if err != nil {
		return err
	}
	return c.redis.Set(ctx, rdb.CacheMetricRequest(request.MetricId), b, c.metricRequestExp).Err()
}

// SetMetricEvExpression save a metric evaluate expression.
func (c *Cache) SetMetricEvExpression(ctx context.Context, metricId int64, expression string) (err error) {
	return c.redis.Set(ctx, rdb.CacheMetricEvExpressionKey(metricId), expression, c.metricEvExpressionExp).Err()
}

// GetMetricEvExpression returns a metric evaluate expression.
func (c *Cache) GetMetricEvExpression(ctx context.Context, metricId int64) (r GetMetricEvExpressionResponse, err error) {
	expression, err := c.redis.Get(ctx, rdb.CacheMetricEvExpressionKey(metricId)).Result()
	if err != nil {
		if err == redis.Nil {
			return r, nil
		} else {
			return r, err
		}
	}
	r.Exists = true
	r.Expression = expression
	return r, err
}

// SetMetricDataPolicyId saves a metric data policy id.
func (c *Cache) SetMetricDataPolicyId(ctx context.Context, metricId int64, dataPolicyId int16) (err error) {
	return c.redis.Set(ctx, rdb.CacheMetricDataPolicyId(metricId), dataPolicyId, c.metricDataPolicyIdExp).Err()
}

// GetMetricDataPolicyId returns a metric data policy id.
func (c *Cache) GetMetricDataPolicyId(ctx context.Context, metricId int64) (r GetMetricDataPolicyIdResponse, err error) {
	id, err := c.redis.Get(ctx, rdb.CacheMetricDataPolicyId(metricId)).Int()
	if err != nil {
		if err == redis.Nil {
			return r, nil
		} else {
			return r, err
		}
	}
	r.Exists = true
	r.DataPolicyId = int16(id)
	return r, nil
}
