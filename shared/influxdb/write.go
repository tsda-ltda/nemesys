package influxdb

import (
	"context"
	"strconv"

	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/types"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/domain"
)

// getMeasurement returns a measurement name for a metric type.
func getMeasurement(mt types.MetricType) string {
	switch mt {
	case types.MTBool:
		return "boolean-metrics"
	case types.MTFloat:
		return "float-metrics"
	case types.MTInt:
		return "integer-metrics"
	default:
		return "integer-metrics"
	}
}

// WritePoint writes a data point into influxdb client buffer.
func (c *Client) WritePoint(ctx context.Context, data models.MetricDataResponse) error {
	if data.Failed {
		return ErrMetricDataResponseIsFailed
	}

	// find bucket
	bucketName := GetBucketName(data.DataPolicyId, false)
	bucket, ok := c.buckets[bucketName]
	if !ok {
		var err error
		bucket, err = c.BucketsAPI().FindBucketByName(ctx, bucketName)
		if err != nil {
			return err
		}
		c.buckets[bucketName] = bucket
	}

	// create point
	p := influxdb2.NewPointWithMeasurement(getMeasurement(data.Type))
	p.AddTag("metric_id", strconv.Itoa(int(data.Id)))
	p.AddField("value", data.Value)

	// write point
	c.WriteAPI(*c.DefaultOrg.Id, *bucket.Id).WritePoint(p)
	return nil
}

func (c *Client) FlushWrites(bucket *domain.Bucket) {

}
