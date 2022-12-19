package influxdb

import (
	"context"
	"fmt"
	"strconv"

	"github.com/fernandotsda/nemesys/shared/env"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

const (
	requestsCountBucketName          = "requests-count"
	requestsCountMeasurement         = "requests"
	requestsRealtimeCountMeasurement = "requests-realtime-data"
	requestsHistoryCountMeasurement  = "requests-history-data"
)

func (c *Client) CreateRequestsCountBucket() (created bool, err error) {
	ctx := context.Background()
	api := c.BucketsAPI()

	bucket, _ := api.FindBucketByName(ctx, requestsCountBucketName)
	if bucket != nil {
		return false, nil
	}

	hours, err := strconv.Atoi(env.RequestsCountBucketRetention)
	if err != nil {
		return false, err
	}

	err = c.createBucket(ctx, requestsCountBucketName, int64(hours*3600))
	return err == nil, err
}

func (c *Client) WriteRequestsCount(count int64) {
	c.writeRequestsCount(count, "requests")
}

func (c *Client) WriteRealtimeDataRequestsCount(count int64) {
	c.writeRequestsCount(count, requestsRealtimeCountMeasurement)
}

func (c *Client) WriteHistoryDataRequestsCount(count int64) {
	c.writeRequestsCount(count, requestsHistoryCountMeasurement)
}

func (c *Client) writeRequestsCount(count int64, measurement string) {
	p := influxdb2.NewPointWithMeasurement(measurement).AddField("count", count)
	c.WriteAPI(*c.DefaultOrg.Id, requestsCountBucketName).WritePoint(p)
}

func (c *Client) GetTotalRequests(ctx context.Context) (total int64, err error) {
	return c.getTotalRequests(ctx, requestsCountMeasurement)
}

func (c *Client) GetTotalRealtimeDataRequests(ctx context.Context) (total int64, err error) {
	return c.getTotalRequests(ctx, requestsRealtimeCountMeasurement)
}

func (c *Client) GetTotalDataHistoryRequests(ctx context.Context) (total int64, err error) {
	return c.getTotalRequests(ctx, requestsHistoryCountMeasurement)
}

func (c *Client) getTotalRequests(ctx context.Context, measurement string) (total int64, err error) {
	bucket, err := c.BucketsAPI().FindBucketByName(ctx, requestsCountBucketName)
	if err != nil {
		return total, err
	}

	if len(bucket.RetentionRules) < 1 {
		return total, ErrInvalidRetentionRulesLength
	}
	retention := bucket.RetentionRules[0].EverySeconds

	table, err := c.QueryAPI(*c.DefaultOrg.Id).Query(context.Background(), getTotalRequestsFlux(measurement, retention))
	if err != nil {
		return total, err
	}
	for table.Next() {
		t, ok := table.Record().Value().(int64)
		if !ok {
			return total, ErrInvalidCountReturn
		}
		total = t
	}
	return total, nil
}

func getTotalRequestsFlux(measurement string, retention int64) string {
	return fmt.Sprintf(`
		from(bucket: "%s")
			|> range(start: -%ds)
			|> filter(fn: (r) => r["_measurement"] == "%s")
			|> filter(fn: (r) => r["_field"] == "count")
			|> sum()
	`, requestsCountBucketName, retention, measurement)
}
