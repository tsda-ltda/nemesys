package dhs

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/fernandotsda/nemesys/shared/env"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/types"
)

func (d *DHS) readDatabase() (err error) {
	ctx := context.Background()

	incialServices, err := strconv.ParseInt(env.InicialDHSServices, 0, 64)
	if err != nil {
		d.log.Fatal("Fail to parse env.InicialDHSServices", logger.ErrField(err))
		return err
	}

	n, err := d.pg.CountFlexLegacyContainers(ctx)
	if err != nil {
		d.log.Fatal("Fail to count flex legacy containers", logger.ErrField(err))
		return
	}
	limit := (n/incialServices)*int64(d.ServiceNumber) + 1
	offset := int64(d.ServiceNumber-1) * limit

	err = d.readFlexLegacyContainers(ctx, int(limit), int(offset))
	if err != nil {
		return err
	}

	n, err = d.pg.CountNonFlexMetrics(ctx)
	if err != nil {
		d.log.Fatal("Fail to count non flex legacy metrics", logger.ErrField(err))
		return
	}
	limit = (n/incialServices)*int64(d.ServiceNumber) + 1
	offset = int64(d.ServiceNumber-1) * limit

	err = d.readMetrics(ctx, int(offset), int(offset))
	if err != nil {
		return err
	}
	return nil
}

func (d *DHS) readFlexLegacyContainers(ctx context.Context, limit int, offset int) (err error) {
	d.log.Info("Reading Flex Legacy containers on database...")
	ids, err := d.pg.GetEnabledContainersIds(ctx, types.CTFlexLegacy, limit, offset)
	if err != nil {
		return err
	}
	for _, id := range ids {
		d.newFlexLegacyPulling(id)
	}
	d.log.Info(fmt.Sprintf("%d flex-legacy containers added", len(ids)))
	return nil
}

func (d *DHS) readMetrics(ctx context.Context, limit int, offset int) (err error) {
	d.log.Info("Reading metrics on database...")
	r, err := d.pg.GetMetricsRequestsAndIntervals(ctx, limit, offset)
	if err != nil {
		return err
	}

	for _, res := range r {
		d.AddMetricPulling(res.MetricRequest, time.Second*time.Duration(res.Interval))
	}
	d.log.Info(fmt.Sprintf("%d metrics added", len(r)))
	return nil
}
