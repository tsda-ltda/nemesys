package influxdb

import (
	"context"
	"fmt"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/domain"
)

const (
	requestsCountBucketName          = "requests-count"
	requestsCountMeasurement         = "requests"
	requestsRealtimeCountMeasurement = "requests-realtime-data"
	requestsHistoryCountMeasurement  = "requests-history-data"

	requestCountRetention = 30 * 24 * 3600 // 30d
)

func (c *Client) CreateRequestsCountBucket() (created bool, err error) {
	ctx := context.Background()
	api := c.BucketsAPI()
	var bucket *domain.Bucket
	defer func() {
		if err == nil {
			c.saveBucketLocal(bucket)
		}
	}()

	bucket, err = api.FindBucketByName(ctx, requestsCountBucketName)
	if bucket != nil {
		return false, nil
	}

	bucket, err = api.CreateBucket(ctx, &domain.Bucket{
		Name: requestsCountBucketName,
		RetentionRules: domain.RetentionRules{{
			EverySeconds: requestCountRetention,
		}},
		OrgID: c.DefaultOrg.Id,
	})
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

func (c *Client) GetTotalRequests() (total int64, err error) {
	return c.getTotalRequests(requestsCountMeasurement)
}

func (c *Client) GetTotalRealtimeDataRequests() (total int64, err error) {
	return c.getTotalRequests(requestsRealtimeCountMeasurement)
}

func (c *Client) GetTotalDataHistoryRequests() (total int64, err error) {
	return c.getTotalRequests(requestsHistoryCountMeasurement)
}

func (c *Client) getTotalRequests(measurement string) (total int64, err error) {
	table, err := c.QueryAPI(*c.DefaultOrg.Id).Query(context.Background(), getTotalRequestsFlux(measurement))
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

func getTotalRequestsFlux(measurement string) string {
	return fmt.Sprintf(`
		from(bucket: "%s")
			|> range(start: -%ds)
			|> filter(fn: (r) => r["_measurement"] == "%s")
			|> filter(fn: (r) => r["_field"] == "count")
			|> sum()
	`, requestsCountBucketName, requestCountRetention, measurement)
}
