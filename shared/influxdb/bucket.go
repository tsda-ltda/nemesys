package influxdb

import (
	"context"

	"github.com/influxdata/influxdb-client-go/v2/domain"
)

func (c *Client) saveBucketLocal(bucket *domain.Bucket) {
	c.buckets[bucket.Name] = bucket
}

func (c *Client) deleteBucketLocal(name string) {
	delete(c.buckets, name)
}

func (c *Client) getBucket(name string) (*domain.Bucket, error) {
	bucket, ok := c.buckets[name]
	if !ok {
		var err error
		bucket, err = c.BucketsAPI().FindBucketByName(context.Background(), name)
		if err != nil {
			return nil, err
		}
		c.saveBucketLocal(bucket)
	}

	return bucket, nil
}

func (c *Client) createBucket(ctx context.Context, name string, retentionSeconds int64) (err error) {
	api := c.BucketsAPI()

	rules := []domain.RetentionRule{{
		EverySeconds: retentionSeconds,
	}}

	bucket, err := api.CreateBucketWithName(ctx, c.DefaultOrg, name, rules...)
	if err != nil {
		return err
	}

	c.saveBucketLocal(bucket)
	return err
}

func (c *Client) updateBucket(ctx context.Context, name string, retentionSeconds int64) (err error) {
	api := c.BucketsAPI()

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
	bucket.RetentionRules = rules

	bucket, err = api.UpdateBucket(ctx, bucket)
	c.saveBucketLocal(bucket)
	return err
}

func (c *Client) deleteBucket(ctx context.Context, name string) (err error) {
	api := c.BucketsAPI()

	bucket, err := api.FindBucketByName(ctx, name)
	if err != nil {
		return err
	}
	c.deleteBucketLocal(name)

	return c.BucketsAPI().DeleteBucket(ctx, bucket)
}
