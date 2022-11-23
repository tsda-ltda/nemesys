package dhs

import (
	"context"
	"fmt"
	"time"
)

func (d *DHS) readDatabase(limit int, offset int) error {
	var round int = offset / limit
	d.log.Info(fmt.Sprintf("(round %d) Reading database...", round))
	r, err := d.pg.GetMetricsRequestsAndIntervals(context.Background(), limit, offset)
	if err != nil {
		return err
	}

	for _, res := range r {
		d.AddMetricPulling(res.MetricRequest, time.Second*time.Duration(res.Interval))
	}
	d.log.Info(fmt.Sprintf("(round %d) Round completed! %d metrics added", round, len(r)))

	if len(r) == limit {
		return d.readDatabase(limit, offset+limit)
	}
	return nil
}
