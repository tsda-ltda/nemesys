package influxdb

import (
	"context"
	"strconv"
	"strings"

	"github.com/fernandotsda/nemesys/shared/types"
)

const defaultQuery = `
	raw = from(bucket: "$raw_bucket")
		|> range(start: $start, stop: $stop)
		|> filter(fn: (r) => r["_measurement"] == "metrics")
		|> filter(fn: (r) => r["metric_id"] == "$metric_id")
		|> filter(fn: (r) => r["_field"] == "$field")
	aggr = from(bucket: "$aggr_bucket")
		|> range(start: $start, stop: $stop)
		|> filter(fn: (r) => r["_measurement"] == "metrics")
		|> filter(fn: (r) => r["metric_id"] == "$metric_id")
		|> filter(fn: (r) => r["_field"] == "$field")
	data = union(tables: [aggr, raw])`

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

	query := replaceQueryVariables(defaultQuery, opts)
	if opts.CustomQueryFlux != "" {
		query += opts.CustomQueryFlux
	} else {
		query += ` data`
	}

	// execute query
	table, err := queryApi.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	points := make([][]any, 0)
	for table.Next() {
		// get record
		r := table.Record()
		points = append(points, []any{r.Time().Unix(), r.Value()})
	}
	return &points, nil
}

func replaceQueryVariables(query string, opts QueryOptions) string {
	query = strings.Replace(query, "$raw_bucket", GetBucketName(opts.DataPolicyId, false), 1)
	query = strings.Replace(query, "$aggr_bucket", GetBucketName(opts.DataPolicyId, true), 1)
	query = strings.Replace(query, "$start", opts.Start, 2)
	query = strings.Replace(query, "$stop", opts.Stop, 2)
	query = strings.Replace(query, "$field", getField(opts.MetricType), 2)
	return strings.Replace(query, "$metric_id", strconv.FormatInt(opts.MetricId, 10), 2)
}
