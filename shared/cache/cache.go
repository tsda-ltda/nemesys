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

	metricRequestByIdentExp   time.Duration
	metricRequestExp          time.Duration
	metricEvExpressionExp     time.Duration
	metricDataPolicyIdExp     time.Duration
	snmpAgentExp              time.Duration
	snmpMetricExp             time.Duration
	customQueryExp            time.Duration
	metricAddDataFormExp      time.Duration
	rtsMetricConfigExp        time.Duration
	serverCostExp             time.Duration
	metricAlarmExpressionsExp time.Duration
	metricAlarmCategoryExp    time.Duration
}

// New returns a prepared Cache struct.
func New() (cache *Cache, err error) {
	// connect to redis
	c, err := rdb.NewCacheClient()
	if err != nil {
		return nil, err
	}

	return &Cache{
		redis:                     c,
		metricRequestByIdentExp:   time.Minute,
		metricRequestExp:          time.Minute,
		metricEvExpressionExp:     time.Minute,
		metricDataPolicyIdExp:     time.Minute,
		snmpAgentExp:              time.Minute * 5,
		snmpMetricExp:             time.Minute * 2,
		customQueryExp:            time.Minute,
		metricAddDataFormExp:      time.Minute * 3,
		rtsMetricConfigExp:        time.Minute * 2,
		serverCostExp:             time.Second * 30,
		metricAlarmExpressionsExp: time.Minute,
		metricAlarmCategoryExp:    time.Minute * 2,
	}, nil
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
