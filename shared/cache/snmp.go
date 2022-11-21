package cache

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/rdb"
	"github.com/go-redis/redis/v8"
)

// GetGoSNMPConfigResponse is the response for the GetGoSNMPConfig handler.
type GetGoSNMPConfigResponse struct {
	// Exists is the config existence.
	Exists bool
	// Agent is the snmp agent.
	Agent models.SNMPAgent
}

// GetSNMPMetricResponse is the response for the GetSNMPMetric handler.
type GetSNMPMetricResponse struct {
	// Exists is the config existence.
	Exists bool
	// Metric is the snmp metric.
	Metric models.SNMPMetric
}

func (c *Cache) SetSNMPAgent(ctx context.Context, containerId int32, agent models.SNMPAgent) (err error) {
	b, err := c.encode(agent)
	if err != nil {
		return err
	}
	return c.Set(ctx, b, rdb.CacheGoSNMPConfig(containerId), c.snmpAgentExp)
}

func (c *Cache) GetSNMPAgent(ctx context.Context, containerId int32) (r GetGoSNMPConfigResponse, err error) {
	b, err := c.Get(ctx, rdb.CacheGoSNMPConfig(containerId))
	if err != nil {
		if err != redis.Nil {
			return r, err
		}
		return r, nil
	}
	r.Exists = true
	err = c.decode(b, &r.Agent)
	return r, err
}

func (c *Cache) SetSNMPMetrics(ctx context.Context, metrics []models.SNMPMetric) (err error) {
	pipe := c.redis.Pipeline()
	for _, m := range metrics {
		b, err := c.encode(m)
		if err != nil {
			return err
		}
		pipe.Set(ctx, rdb.CacheSNMPMetric(m.Id), b, c.snmpMetricExp)
	}
	_, err = pipe.Exec(ctx)
	return err
}

func (c *Cache) GetSNMPMetrics(ctx context.Context, ids []int64) (r []GetSNMPMetricResponse, err error) {
	pipe := c.redis.Pipeline()
	cmds := make([]*redis.StringCmd, len(ids))
	for i, id := range ids {
		cmds[i] = pipe.Get(ctx, rdb.CacheSNMPMetric(id))
	}
	_, err = pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return r, err
	}

	r = make([]GetSNMPMetricResponse, len(ids))
	for i, cmd := range cmds {
		metric := models.SNMPMetric{
			Id: ids[i],
		}

		b, err := cmd.Bytes()
		if err != nil {
			if err != redis.Nil {
				return r, err
			}
			r[i].Metric = metric
			continue
		}
		err = c.decode(b, &metric)
		if err != nil {
			return r, err
		}
		r[i].Exists = true
		r[i].Metric = metric
	}
	return r, nil
}
