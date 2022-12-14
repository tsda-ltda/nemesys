package influxdb

import (
	"context"
	"strconv"
	"time"

	"github.com/fernandotsda/nemesys/shared/env"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/domain"
)

const (
	logsBucketName = "logs"
)

// CreateLogsBucket creates the logs bucket if not exists.
func (c *Client) CreateLogsBucket() (created bool, err error) {
	ctx := context.Background()
	api := c.BucketsAPI()

	bucket, _ := api.FindBucketByName(ctx, logsBucketName)
	if bucket != nil {
		return false, nil
	}

	hours, err := strconv.Atoi(env.LogsBucketRetention)
	if err != nil {
		return false, err
	}

	err = c.createBucket(ctx, logsBucketName, int64(hours*3600))
	return err == nil, err
}

func (c *Client) UpdateLogsBucket(retentionHours int64) (err error) {
	ctx := context.Background()
	api := c.BucketsAPI()
	var bucket *domain.Bucket
	defer func() {
		if err == nil {
			c.saveBucketLocal(bucket)
		}
	}()

	bucket, err = api.FindBucketByName(ctx, logsBucketName)
	if err != nil {
		return err
	}

	rules := bucket.RetentionRules
	if len(rules) == 0 {
		return ErrInvalidRetentionRulesLength
	}
	rule := rules[0]
	rule.EverySeconds = retentionHours * 60
	rules[0] = rule
	bucket.RetentionRules = rules
	bucket, err = api.UpdateBucket(ctx, bucket)
	return err
}

func (c *Client) WriteLog(ctx context.Context, log map[string]any) error {
	if log == nil {
		return ErrLogIsNil
	}

	bucket, err := c.getBucket(logsBucketName)
	if err != nil {
		return err
	}

	level, ok1 := log["level"].(string)
	serv, ok2 := log["serv"].(string)
	ts, ok3 := log["ts"].(string)
	timestamp, err := time.Parse(time.RFC3339Nano, ts)
	msg := log["msg"]

	if !ok1 || !ok2 || !ok3 || err != nil || msg == nil {
		return ErrInvalidLog
	}

	p := influxdb2.NewPointWithMeasurement(logsBucketName)
	p.AddTag("level", level)
	p.AddTag("serv", serv)
	p.SetTime(timestamp)

	for name, v := range log {
		if name == "level" || name == "ts" || name == "serv" {
			continue
		}
		p.AddField(name, v)
	}

	c.WriteAPI(*c.DefaultOrg.Id, *bucket.Id).WritePoint(p)
	return nil
}
