package influxdb

import (
	"context"
	"fmt"

	"github.com/fernandotsda/nemesys/shared/models"
)

func GetBucketName(dataPolicyId int16, aggr bool) string {
	var suffix string
	if aggr {
		suffix = "aggr"
	} else {
		suffix = "raw"
	}
	return fmt.Sprintf("%d-%s", dataPolicyId, suffix)
}

// CreateDataPolicy creates two buckets on database to represent a data policy.
func (c *Client) CreateDataPolicy(ctx context.Context, dp models.DataPolicy) (err error) {
	err = c.createBucket(ctx, GetBucketName(dp.Id, false), int64(dp.Retention*3600))
	if err != nil {
		return err
	}

	if dp.UseAggr {
		err = c.createBucket(ctx, GetBucketName(dp.Id, true), int64(dp.AggrRetention*3600+dp.Retention*3600))
		if err != nil {
			return err
		}
		err = c.createAggrTask(ctx, dp)
	}

	return err
}

// UpdateDataPolicy updates/removes the data policy buckets according to the data policy update.
func (c *Client) UpdateDataPolicy(ctx context.Context, dp models.DataPolicy) (err error) {
	api := c.BucketsAPI()

	err = c.updateBucket(ctx, GetBucketName(dp.Id, false), int64(dp.Retention*3600))
	if err != nil {
		return err
	}

	aggrName := GetBucketName(dp.Id, true)
	aggrRetention := int64(dp.AggrRetention*3600 + dp.Retention*3600)

	_, err = api.FindBucketByName(ctx, aggrName)
	if dp.UseAggr {
		if err != nil {
			err = c.createBucket(ctx, aggrName, aggrRetention)
			if err != nil {
				return err
			}
			err = c.createAggrTask(ctx, dp)
			if err != nil {
				return err
			}
		} else {
			err = c.updateBucket(ctx, aggrName, aggrRetention)
			if err != nil {
				return err
			}
			err = c.updateAggrTask(ctx, dp)
			if err != nil {
				return err
			}
		}
	} else {
		if err == nil {
			err = c.deleteBucket(ctx, aggrName)
			if err != nil {
				return err
			}
			err = c.deleteAggrTask(ctx, dp.Id)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// DeleteDataPolicy deletes the data policy buckets.
func (c *Client) DeleteDataPolicy(ctx context.Context, id int16) (err error) {
	api := c.BucketsAPI()

	err = c.deleteBucket(ctx, GetBucketName(id, false))
	if err != nil {
		return err
	}
	_, err = api.FindBucketByName(ctx, GetBucketName(id, true))
	if err == nil {
		err = c.deleteAggrTask(ctx, id)
		if err != nil {
			return err
		}
		err = c.deleteBucket(ctx, GetBucketName(id, true))
		if err != nil {
			return err
		}
	}
	return nil
}
