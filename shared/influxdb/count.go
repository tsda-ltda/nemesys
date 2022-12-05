package influxdb

import (
	"context"
	"fmt"

	"github.com/fernandotsda/nemesys/shared/models"
)

func (c *Client) CountAllMetricsPoints(dps []models.DataPolicy) (n int, err error) {
	if len(dps) == 0 {
		return 0, nil
	}
	query := ""
	var tables string
	for i, dp := range dps {
		query += fmt.Sprintf(`
			raw_%d = from(bucket: "%s")
				|> range(start: -%dh)
				|> filter(fn: (r) => r["_measurement"] == "metrics")
				|> count()
			aggr_%d = from(bucket: "%s")
				|> range(start: -%dh)
				|> filter(fn: (r) => r["_measurement"] == "metrics")
				|> count()`,
			dp.Id, GetBucketName(dp.Id, false), dp.Retention, dp.Id, GetBucketName(dp.Id, true), dp.AggrRetention)
		if i != 0 {
			tables += ","
		}
		tables += fmt.Sprintf(`raw_%d, aggr_%d`, dp.Id, dp.Id)
	}
	query += fmt.Sprintf(`data = union(tables: [%s]) data`, tables)

	table, err := c.QueryAPI(*c.DefaultOrg.Id).Query(context.Background(), query)
	if err != nil {
		return n, err
	}
	for table.Next() {
		v := table.Record().Value()
		_n, ok := v.(int64)
		if !ok {
			return n, ErrInvalidCountReturn
		}
		n += int(_n)
	}
	return n, nil
}
