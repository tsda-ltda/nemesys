package cache

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/rdb"
	"github.com/go-redis/redis/v8"
)

// GetMetricRequestByIdentResponse is the response for the GetMetricRequestByIdent handler.
type GetMetricRequestByIdentResponse struct {
	// Exists is the cache existence.
	Exists bool
	// Request is the metric request.
	Request models.MetricRequest
}

// GetMetricEvExpressionResponse is the response for the GetMetricEvExpression handler.
type GetMetricEvExpressionResponse struct {
	// Exists is the cache existence.
	Exists bool
	// Expression is the metric evaluate expression.
	Expression string
}

// GetMetricDataPolicyIdResponse is the respose for GetMetricDataPolicyId handler.
type GetMetricDataPolicyIdResponse struct {
	// Exists is the cache existence.
	Exists bool
	// DataPolicyId is the data policy id
	DataPolicyId int16
}

// GetMetricRequestByIdent returns a metric request information.
func (c *Cache) GetMetricRequestByIdent(ctx context.Context, teamIdent string, contextIdent string, metricIdent string) (r GetMetricRequestByIdentResponse, err error) {
	bytes, err := c.redis.Get(ctx, rdb.RDBCacheMetricIdContainerIdKey(teamIdent, contextIdent, metricIdent)).Bytes()
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
	return c.redis.Set(ctx, rdb.RDBCacheMetricIdContainerIdKey(teamIdent, contextIdent, metricIdent), bytes, c.metricIdContainerIdExp).Err()
}

// SetMetricEvExpression save a metric evaluate expression.
func (c *Cache) SetMetricEvExpression(ctx context.Context, metricId int64, expression string) (err error) {
	return c.redis.Set(ctx, rdb.RDBCacheMetricEvExpressionKey(metricId), expression, c.metricEvExpressionExp).Err()
}

// GetMetricEvExpression returns a metric evaluate expression.
func (c *Cache) GetMetricEvExpression(ctx context.Context, metricId int64) (r GetMetricEvExpressionResponse, err error) {
	expression, err := c.redis.Get(ctx, rdb.RDBCacheMetricEvExpressionKey(metricId)).Result()
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
	return c.redis.Set(ctx, rdb.RDBCacheMetricDataPolicyId(metricId), dataPolicyId, c.metricDataPolicyIdExp).Err()
}

// GetMetricDataPolicyId returns a metric data policy id.
func (c *Cache) GetMetricDataPolicyId(ctx context.Context, metricId int64) (r GetMetricDataPolicyIdResponse, err error) {
	id, err := c.redis.Get(ctx, rdb.RDBCacheMetricDataPolicyId(metricId)).Int()
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
