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
	// DataPolicyId is the data policy id.
	DataPolicyId int16
}

type GetMetricAddDataFormResponse struct {
	// Exists is the cache existence.
	Exists bool
	// Form is the base metric add data form.
	Form models.BasicMetricAddDataForm
}

type GetRTSMetricConfigResponse struct {
	// Exists is the cache existence.
	Exists bool
	// Config is the real time service metric configuration.
	Config models.RTSMetricConfig
}

func (c *Cache) GetMetricAddDataForm(ctx context.Context, refkey string) (r GetMetricAddDataFormResponse, err error) {
	b, err := c.Get(ctx, rdb.CacheMetricAddDataFormKey(refkey))
	if err != nil {
		if err == redis.Nil {
			return r, nil
		}
		return r, err
	}
	err = c.decode(b, &r.Form)
	r.Exists = true
	return r, err
}

func (c *Cache) SetMetricAddDataForm(ctx context.Context, refkey string, form models.BasicMetricAddDataForm) (err error) {
	b, err := c.encode(form)
	if err != nil {
		return err
	}
	return c.redis.Set(ctx, rdb.CacheMetricAddDataFormKey(refkey), b, c.metricAddDataFormExp).Err()
}

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

func (c *Cache) SetMetricEvExpression(ctx context.Context, metricId int64, expression string) (err error) {
	return c.redis.Set(ctx, rdb.CacheMetricEvExpressionKey(metricId), expression, c.metricEvExpressionExp).Err()
}

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

func (c *Cache) SetMetricDataPolicyId(ctx context.Context, metricId int64, dataPolicyId int16) (err error) {
	return c.redis.Set(ctx, rdb.CacheMetricDataPolicyIdKey(metricId), dataPolicyId, c.metricDataPolicyIdExp).Err()
}

func (c *Cache) GetMetricDataPolicyId(ctx context.Context, metricId int64) (r GetMetricDataPolicyIdResponse, err error) {
	id, err := c.redis.Get(ctx, rdb.CacheMetricDataPolicyIdKey(metricId)).Int()
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

func (c *Cache) SetRTSMetricConfig(ctx context.Context, metricId int64, config models.RTSMetricConfig) (err error) {
	b, err := c.encode(config)
	if err != nil {
		return err
	}
	return c.redis.Set(ctx, rdb.CacheRTSMetricConfig(metricId), b, c.rtsMetricConfigExp).Err()
}

func (c *Cache) GetRTSMetricConfig(ctx context.Context, metricId int64) (r GetRTSMetricConfigResponse, err error) {
	b, err := c.Get(ctx, rdb.CacheRTSMetricConfig(metricId))
	if err != nil {
		if err == redis.Nil {
			return r, nil
		}
		return r, err
	}
	err = c.decode(b, &r.Config)
	r.Exists = true
	return r, err
}
