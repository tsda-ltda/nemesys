package influxdb

import (
	"context"
	"fmt"
	"strconv"

	"github.com/fernandotsda/nemesys/shared/env"
	"github.com/fernandotsda/nemesys/shared/models"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

const (
	alarmHistoryBucketName      = "alarm-history"
	alarmHistoryMeasurementName = "history"
)

type QueryAlarmHistoryOptions struct {
	// Start is the range start in seconds.
	Start int64
	// Stop is the range stop in seconds.
	Stop int64
	// ContianerId is the container identifier.
	ContainerId int32
	// MetricId is the metric identfier.
	MetricId int64
	// Level is the alarm category level. Nil value
	// means get from all levels
	Level *int32
}

// CreateAlarmHistoryBucket creates an alarm history bucket with
// the retention configurated in the enviroment.
func (c *Client) CreateAlarmHistoryBucket() (cteated bool, err error) {
	ctx := context.Background()
	api := c.BucketsAPI()

	bucket, _ := api.FindBucketByName(ctx, alarmHistoryBucketName)
	if bucket != nil {
		return false, nil
	}

	hours, err := strconv.Atoi(env.AlarmHistoryBucketRetention)
	if err != nil {
		return false, err
	}

	err = c.createBucket(context.Background(), alarmHistoryBucketName, int64(hours*3600))
	return err == nil, err
}

// UpdateAlarmHistoryBucket updates the retention of the alarm-history
// bucket.
func (c *Client) UpdateAlarmHistoryBucket(ctx context.Context, hours int64) (err error) {
	return c.updateBucket(ctx, alarmHistoryBucketName, hours*3600)
}

// WriteAlarmOcurrency writes a point on influx client buffer.
func (c *Client) WriteAlarmOccurency(occurency models.AlarmOccurency) {
	p := influxdb2.NewPointWithMeasurement(alarmHistoryMeasurementName)

	p.SetTime(occurency.Time)
	p.AddTag("container_id", strconv.FormatInt(int64(occurency.ContainerId), 10))

	p.AddField("metric_id", occurency.MetricId)
	p.AddField("value", occurency.Value)
	p.AddField("level", occurency.Category.Level)

	c.WriteAPI(*c.DefaultOrg.Id, alarmHistoryBucketName).WritePoint(p)
}

func (c *Client) QueryAlarmHistory(ctx context.Context, options QueryAlarmHistoryOptions) (points [][3]any, err error) {
	api := c.QueryAPI(*c.DefaultOrg.Id)

	query := getAlarmHistoryQuery(options)
	table, err := api.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	points = make([][3]any, 0)
	for table.Next() {
		r := table.Record()
		points = append(points, [3]any{r.Time().Unix(), r.ValueByKey("value"), r.ValueByKey("level")})
	}
	return points, nil
}

func getAlarmHistoryQuery(options QueryAlarmHistoryOptions) string {
	var startS string = strconv.FormatInt(options.Start, 10)
	var stopS string

	if options.Stop == 0 {
		stopS = "now()"
	} else {
		stopS = strconv.FormatInt(options.Stop, 10)
	}

	if options.Level == nil {
		return fmt.Sprintf(`from(bucket: "%s")
		|> range(start: %s, stop: %s)
			|> filter(fn: (r) => r["_measurement"] == "%s")
			|> filter(fn: (r) => r["container_id"] == "%d")
			|> filter(fn: (r) => r["_field"] == "metric_id" or r["_field"] == "level" or r["_field"] == "value")
			|> pivot(rowKey:["_time"], columnKey:["_field"], valueColumn:"_value")
			|> filter(fn: (r) => r.metric_id == %d)`,
			alarmHistoryBucketName,
			startS,
			stopS,
			alarmHistoryMeasurementName,
			options.ContainerId,
			options.MetricId,
		)
	}
	return fmt.Sprintf(`from(bucket: "%s")
		  	|> range(start: %s, stop: %s)
			  |> filter(fn: (r) => r["_measurement"] == "%s")
			  |> filter(fn: (r) => r["container_id"] == "%d")
			  |> filter(fn: (r) => r["_field"] == "metric_id" or r["_field"] == "level" or r["_field"] == "value")
			  |> pivot(rowKey:["_time"], columnKey:["_field"], valueColumn:"_value")
			  |> filter(fn: (r) => r.metric_id == %d and r.level == %d)`,
		alarmHistoryBucketName,
		startS,
		stopS,
		alarmHistoryMeasurementName,
		options.ContainerId,
		options.MetricId,
		*options.Level,
	)
}
