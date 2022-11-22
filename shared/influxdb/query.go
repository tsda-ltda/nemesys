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
	// Start is the start range. Can be ommited.
	Start string
	// Stop is the end range. Can be ommited.
	Stop string
	// CustomQueryFlux is the custom query flux. Can be ommited.
	CustomQueryFlux string
	// DataPolicyId is the data policy id.
	DataPolicyId int16
	// Metric id is the metric id.
	MetricId int64
	// MetricType is the metric type.
	MetricType types.MetricType
}

func (c *Client) Query(ctx context.Context, opts QueryOptions) (queryPoints *[][]any, err error) {
	queryApi := c.QueryAPI(*c.DefaultOrg.Id)

	rawBucket, err := c.getBucketLocal(GetBucketName(opts.DataPolicyId, false))
	if err != nil {
		return nil, err
	}

	query, err := getBaseQuery(opts, rawBucket)
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

	points := make([][]any, 0)
	for table.Next() {
		r := table.Record()
		points = append(points, []any{r.Time().Unix(), r.Value()})
	}
	return &points, nil
}

func getBaseQuery(opts QueryOptions, rawBucket *domain.Bucket) (query string, err error) {
	if len(rawBucket.RetentionRules) == 0 {
		return query, errors.New("invalid bucket")
	}

	retentionSeconds := rawBucket.RetentionRules[0].EverySeconds
	retention := time.Duration(retentionSeconds) * time.Second
	start, err := ParseDuration(opts.Start)
	if err != nil {
		return query, err
	}

	if retention+start >= 0 {
		query = fmt.Sprintf(`
			data = from(bucket: "%s")
				|> range(start: %s, stop: %s)
				|> filter(fn: (r) => r["_measurement"] == "metrics")
				|> filter(fn: (r) => r["metric_id"] == "%s")
				|> filter(fn: (r) => r["_field"] == "%s")`,
			GetBucketName(opts.DataPolicyId, false),
			opts.Start,
			opts.Stop,
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
			opts.Stop,
			strconv.FormatInt(opts.MetricId, 10),
			getField(opts.MetricType),
			GetBucketName(opts.DataPolicyId, true),
			opts.Start,
			DurationFromSeconds(-retentionSeconds),
			strconv.FormatInt(opts.MetricId, 10),
			getField(opts.MetricType),
		)
	}
	return query, nil
}
