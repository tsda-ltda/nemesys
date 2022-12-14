package influxdb

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/fernandotsda/nemesys/shared/types"
	"github.com/influxdata/influxdb-client-go/v2/domain"
)

type QueryOptions struct {
	// Start is the start range in seconds.
	Start int64
	// Stop is the end range in seconds. Can be ommited.
	Stop int64
	// CustomQueryFlux is the custom query flux. Can be ommited.
	CustomQueryFlux string
	// DataPolicyId is the data policy id.
	DataPolicyId int16
	// Metric id is the metric id.
	MetricId int64
	// MetricType is the metric type.
	MetricType types.MetricType
}

func (c *Client) Query(ctx context.Context, opts QueryOptions) (points [][]any, err error) {
	queryApi := c.QueryAPI(*c.DefaultOrg.Id)

	rawBucket, err := c.getBucket(GetBucketName(opts.DataPolicyId, false))
	if err != nil {
		return nil, err
	}
	_, err = c.getBucket(GetBucketName(opts.DataPolicyId, true))

	query, err := getBaseQuery(opts, rawBucket, err == nil)
	if err != nil {
		return nil, ErrInvalidQueryOptions
	}

	if opts.CustomQueryFlux != "" {
		query += opts.CustomQueryFlux
	} else {
		query += ` data`
	}

	table, err := queryApi.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	points = make([][]any, 0)
	for table.Next() {
		r := table.Record()
		points = append(points, []any{r.Time().Unix(), r.Value()})
	}
	return points, nil
}

func getBaseQuery(opts QueryOptions, rawBucket *domain.Bucket, useAggr bool) (query string, err error) {
	if len(rawBucket.RetentionRules) == 0 {
		return query, errors.New("invalid bucket")
	}

	var stopDur string
	var startDur string

	if opts.Stop == 0 {
		stopDur = "now()"
	} else {
		stopDur = strconv.FormatInt(opts.Stop, 10)
	}
	startDur = strconv.FormatInt(opts.Start, 10)

	retentionSeconds := rawBucket.RetentionRules[0].EverySeconds
	retention := time.Duration(retentionSeconds) * time.Second

	if retention+(time.Duration(opts.Start)*time.Second) >= 0 || !useAggr {
		query = fmt.Sprintf(`
			data = from(bucket: "%s")
				|> range(start: %s, stop: %s)
				|> filter(fn: (r) => r["_measurement"] == "metrics")
				|> filter(fn: (r) => r["metric_id"] == "%s")
				|> filter(fn: (r) => r["_field"] == "%s")`,
			GetBucketName(opts.DataPolicyId, false),
			startDur,
			stopDur,
			strconv.FormatInt(opts.MetricId, 10),
			getField(opts.MetricType),
		)
	} else {
		query = fmt.Sprintf(`
			raw = from(bucket: "%s")
				|> range(start: %s, stop: %s)
				|> filter(fn: (r) => r["_measurement"] == "metrics")
				|> filter(fn: (r) => r["metric_id"] == "%s")
				|> filter(fn: (r) => r["_field"] == "%s")
			aggr = from(bucket: "%s")
				|> range(start: %s, stop: %s)
				|> filter(fn: (r) => r["_measurement"] == "metrics")
				|> filter(fn: (r) => r["metric_id"] == "%s")
				|> filter(fn: (r) => r["_field"] == "%s")
			data = union(tables: [aggr, raw])`,
			GetBucketName(opts.DataPolicyId, false),
			DurationFromSeconds(-retentionSeconds),
			stopDur,
			strconv.FormatInt(opts.MetricId, 10),
			getField(opts.MetricType),
			GetBucketName(opts.DataPolicyId, true),
			startDur,
			DurationFromSeconds(-retentionSeconds),
			strconv.FormatInt(opts.MetricId, 10),
			getField(opts.MetricType),
		)
	}
	return query, nil
}
