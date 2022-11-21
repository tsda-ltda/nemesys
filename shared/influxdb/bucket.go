package influxdb

import (
	"context"

	"github.com/influxdata/influxdb-client-go/v2/domain"
)

func (c *Client) createBucket(ctx context.Context, name string, description string, retentionSeconds int64) (err error) {
	api := c.BucketsAPI()

	rules := []domain.RetentionRule{{
		EverySeconds: retentionSeconds,
	}}

	// create bucket
	bucket, err := api.CreateBucketWithName(ctx, c.DefaultOrg, name, rules...)
	if err != nil {
		return err
	}

	// update description
	bucket.Description = fmtBucketDescription(description)
	bucket, err = api.UpdateBucket(ctx, bucket)
	c.buckets[name] = bucket
	return err
}

func (c *Client) updateBucket(ctx context.Context, name string, description string, retentionSeconds int64) (err error) {
	api := c.BucketsAPI()

	// find bucket
	bucket, err := api.FindBucketByName(ctx, name)
	if err != nil {
		return err
	}

	var shardGroupDuration int64
	if retentionSeconds < 172800 {
		shardGroupDuration = 3600 // 1h
	} else if retentionSeconds < 15552000 {
		shardGroupDuration = 86400 // 1d
	} else {
		shardGroupDuration = 604800 // 7d
	}

	rules := []domain.RetentionRule{{
		EverySeconds:              retentionSeconds,
		ShardGroupDurationSeconds: &shardGroupDuration,
	}}

	// update params
	bucket.RetentionRules = rules
	bucket.Description = &description

	// update bucket
	bucket, err = api.UpdateBucket(ctx, bucket)
	c.buckets[name] = bucket
	return err
}

func (c *Client) deleteBucket(ctx context.Context, name string) (err error) {
	api := c.BucketsAPI()

	// find bucket
	bucket, err := api.FindBucketByName(ctx, name)
	if err != nil {
		return err
	}
	delete(c.buckets, name)

	// delete bucket
	return c.BucketsAPI().DeleteBucket(ctx, bucket)
}
