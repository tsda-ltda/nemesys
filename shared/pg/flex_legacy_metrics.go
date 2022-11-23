package pg

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/models"
)

type FlexLegacyMetricGetResponse struct {
	// Exists is the metric existence.
	Exists bool
	// Metric is the metric.
	Metric models.Metric[models.FlexLegacyMetric]
}

type FlexLegacyMetricGetProtocolResponse struct {
	// Exists is the metric existence.
	Exists bool
	// Metric is the metric.
	Metric models.FlexLegacyMetric
}

type FlexLegacyMetricsGetAsSNMPMetricResponse struct {
	// Exists is the metric existence.
	Exists bool
	// Metric is the metric.
	Metric models.SNMPMetric
}

type FlexLegacyMetricsGetByIdsAsSNMPMetricResponse struct {
	// Exists is the metric existence.
	Exists bool
	// Metric is the metric.
	Metric models.SNMPMetric
}

const (
	sqlFlexLegacyMetricsCreate               = `INSERT INTO flex_legacy_metrics (metric_id, oid, port, port_type) VALUES($1, $2, $3, $4);`
	sqlFlexLegacyMetricsUpdate               = `UPDATE flex_legacy_metrics SET (oid, port, port_type) = ($2, $3, $4) WHERE metric_id = $1;`
	sqlFlexLegacyMetricsGetProtocol          = `SELECT oid, port, port_type FROM flex_legacy_metrics WHERE metric_id = $1;`
	sqlFlexLegacyMetricsGetAsSNMPMetric      = `SELECT oid FROM flex_legacy_metrics WHERE metric_id = $1;`
	sqlFlexLegacyMetricsGetByIdsAsSNMPMetric = `SELECT metric_id, oid FROM flex_legacy_metrics WHERE metric_id = ANY ($1);`
)

func (pg *PG) CreateFlexLegacyMetric(ctx context.Context, metric models.Metric[models.FlexLegacyMetric]) (err error) {
	c, err := pg.pool.Begin(ctx)
	if err != nil {
		return err
	}
	id, err := pg.createMetric(ctx, c, metric.Base)
	if err != nil {
		c.Rollback(ctx)
		return err
	}
	_, err = c.Exec(ctx, sqlFlexLegacyMetricsCreate,
		id,
		metric.Protocol.OID,
		metric.Protocol.Port,
		metric.Protocol.PortType,
	)
	if err != nil {
		c.Rollback(ctx)
		return err
	}
	return c.Commit(ctx)
}

func (pg *PG) UpdateFlexLegacyMetric(ctx context.Context, metric models.Metric[models.FlexLegacyMetric]) (exists bool, err error) {
	c, err := pg.pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	exists, err = pg.updateMetric(ctx, c, metric.Base)
	if err != nil {
		c.Rollback(ctx)
		return false, err
	}
	if !exists {
		return false, nil
	}
	t, err := c.Exec(ctx, sqlFlexLegacyMetricsUpdate,
		metric.Base.Id,
		metric.Protocol.OID,
		metric.Protocol.Port,
		metric.Protocol.PortType,
	)
	if err != nil {
		c.Rollback(ctx)
		return false, err
	}
	return t.RowsAffected() != 0, c.Commit(ctx)
}

func (pg *PG) GetFlexLegacyMetric(ctx context.Context, id int64) (r FlexLegacyMetricGetResponse, err error) {
	baseR, err := pg.GetMetric(ctx, id)
	if err != nil {
		return r, err
	}
	protocolR, err := pg.GetFlexLegacyMetricProtocol(ctx, id)
	if err != nil {
		return r, err
	}
	r.Exists = baseR.Exists
	r.Metric.Base = baseR.Metric
	r.Metric.Protocol = protocolR.Metric
	return r, nil
}

func (pg *PG) GetFlexLegacyMetricProtocol(ctx context.Context, id int64) (r FlexLegacyMetricGetProtocolResponse, err error) {
	c, err := pg.pool.Acquire(ctx)
	if err != nil {
		return r, err
	}
	defer c.Release()
	rows, err := c.Query(ctx, sqlFlexLegacyMetricsGetProtocol, id)
	if err != nil {
		return r, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(
			&r.Metric.OID,
			&r.Metric.Port,
			&r.Metric.PortType,
		)
		if err != nil {
			return r, err
		}
		r.Metric.Id = id
		r.Exists = true
	}
	return r, nil
}

func (pg *PG) DeleteFlexLegacyMetric(ctx context.Context, id int64) (exists bool, err error) {
	return pg.DeleteMetric(ctx, id)
}

func (pg *PG) GetFlexLegacyMetricAsSNMPMetric(ctx context.Context, id int64) (r FlexLegacyMetricsGetAsSNMPMetricResponse, err error) {
	c, err := pg.pool.Acquire(ctx)
	if err != nil {
		return r, err
	}
	defer c.Release()
	rows, err := c.Query(ctx, sqlFlexLegacyMetricsGetAsSNMPMetric, id)
	if err != nil {
		return r, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(
			&r.Metric.OID,
		)
		if err != nil {
			return r, err
		}
		r.Metric.Id = id
		r.Exists = true
	}
	return r, nil
}

func (pg *PG) FlexLegacyMetricsByIdsAsSNMPMetric(ctx context.Context, ids []int64) (metrics []models.SNMPMetric, err error) {
	c, err := pg.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Release()
	rows, err := c.Query(ctx, sqlFlexLegacyMetricsGetByIdsAsSNMPMetric, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var m models.SNMPMetric
		err = rows.Scan(
			&m.Id,
			&m.OID,
		)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, m)
	}
	return metrics, nil
}
