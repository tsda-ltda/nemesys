package cache

import (
	"context"
	"time"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/rdb"
	"github.com/go-redis/redis/v8"
)

type Cache struct {
	// redis is the redis client used to get and set data.
	redis *redis.Client

	// metricRequestByIdentExp is the time to expire the metric request by ident.
	metricRequestByIdentExp time.Duration
	// metricRequestExp is the time to expire the metric request.
	metricRequestExp time.Duration
	// metricEvExpressionExp is the time to expire the evaluate expression.
	metricEvExpressionExp time.Duration
	// metricEvExpressionExp is the time to expire the data policy id.
	metricDataPolicyIdExp time.Duration
	// snmpAgentExp is the time to expire the snmp agent.
	snmpAgentExp time.Duration
	// snmpMetricExp is the time to expire the snmp metric.
	snmpMetricExp time.Duration
	// customQueryExp is the time to expire the custom query.
	customQueryExp time.Duration
}

// New returns a prepared Cache struct.
func New() *Cache {
	// connect to redis
	c, err := rdb.NewCacheClient()
	if err != nil {
		panic("fail to connect to redis cache database")
	}

	return &Cache{
		redis:                   c,
		metricRequestByIdentExp: time.Minute,
		metricRequestExp:        time.Minute,
		metricEvExpressionExp:   time.Minute,
		metricDataPolicyIdExp:   time.Minute,
		snmpAgentExp:            time.Minute * 5,
		snmpMetricExp:           time.Minute * 2,
		customQueryExp:          time.Minute,
	}
}

func (c *Cache) Close() {
	c.redis.Close()
}

func (c *Cache) encode(v any) ([]byte, error) {
	return amqp.Encode(v)
}

func (c *Cache) decode(b []byte, v any) error {
	return amqp.Decode(b, v)
}

func (c *Cache) Set(ctx context.Context, b []byte, key string, exp time.Duration) error {
	return c.redis.Set(ctx, key, b, exp).Err()
}

func (c *Cache) Get(ctx context.Context, key string) (b []byte, err error) {
	return c.redis.Get(ctx, key).Bytes()
}
