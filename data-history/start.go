package dhs

import (
	"context"
	"fmt"
	"time"

	"github.com/fernandotsda/nemesys/shared/types"
)

func (d *DHS) readDatabase() (err error) {
	err = d.readFlexLegacyContainers(250, 0)
	if err != nil {
		return err
	}
	err = d.readMetrics(100, 0)
	if err != nil {
		return err
	}
	return nil
}

func (d *DHS) readFlexLegacyContainers(limit int, offset int) (err error) {
	var round int = offset / limit
	d.log.Info(fmt.Sprintf("(round %d) Reading Flex Legacy containers on database...", round))
	ids, err := d.pg.GetEnabledContainersIds(context.Background(), types.CTFlexLegacy, limit, offset)
	if err != nil {
		return err
	}
	for _, id := range ids {
		d.newFlexLegacyPulling(id)
	}
	d.log.Info(fmt.Sprintf("(round %d) Round completed! %d flex-legacy containers added", round, len(ids)))

	if len(ids) == limit {
		return d.readFlexLegacyContainers(limit, offset+limit)
	}
	return nil
}

func (d *DHS) readMetrics(limit int, offset int) (err error) {
	var round int = offset / limit
	d.log.Info(fmt.Sprintf("(round %d) Reading metrics on database...", round))
	r, err := d.pg.GetMetricsRequestsAndIntervals(context.Background(), limit, offset)
	if err != nil {
		return err
	}

	for _, res := range r {
		d.AddMetricPulling(res.MetricRequest, time.Second*time.Duration(res.Interval))
	}
	d.log.Info(fmt.Sprintf("(round %d) Round completed! %d metrics added", round, len(r)))

	if len(r) == limit {
		return d.readMetrics(limit, offset+limit)
	}
	return nil
}
