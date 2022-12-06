package influxdb

import (
	"context"
	"strconv"
	"time"

	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/types"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/domain"
)

// getField returns a field name for a metric type.
func getField(mt types.MetricType) string {
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

// WritePoint writes a data point into influxdb client buffer. ContainerId is ignored.
func (c *Client) WritePoint(ctx context.Context, data models.MetricDataResponse, timestamp time.Time) error {
	if data.Failed {
		return ErrMetricDataResponseIsFailed
	}

	// find bucket
	bucketName := GetBucketName(data.DataPolicyId, false)
	bucket, err := c.getBucket(bucketName)
	if err != nil {
		return err
	}

	// create point
	p := influxdb2.NewPointWithMeasurement("metrics")
	p.AddTag("metric_id", strconv.Itoa(int(data.Id)))
	p.AddField(getField(data.Type), data.Value)
	p.SetTime(timestamp)

	// write point
	c.WriteAPI(*c.DefaultOrg.Id, *bucket.Id).WritePoint(p)
	return nil
}

func (c *Client) FlushWrites(bucket *domain.Bucket) {

}
