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
func (c *Client) WritePoint(ctx context.Context, data models.MetricDataResponse, bucket *domain.Bucket) {
	if data.Failed {
		return
	}

	api := c.WriteAPI(*c.DefaultOrg.Id, *bucket.Id)

	// create point
	p := influxdb2.NewPointWithMeasurement(getMeasurement(data.Type))
	p.AddTag("container", strconv.Itoa(int(data.ContainerId)))
	p.AddField("id", data.Id)
	p.AddField("value", data.Value)

	// write point
	api.WritePoint(p)
}

func (c *Client) FlushWrites(bucket *domain.Bucket) {

}
