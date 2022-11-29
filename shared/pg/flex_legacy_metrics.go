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
	sqlFlexLegacyMetricsGetMetricsRequests   = `SELECT
		m.id, m.type, m.data_policy_id, f.port, f.port_type 
		FROM metrics m FULL JOIN flex_legacy_metrics f ON m.id = f.metric_id
		WHERE m.enabled = true AND m.container_id = $1 AND m.dhs_enabled = true;`
)

func (pg *PG) CreateFlexLegacyMetric(ctx context.Context, metric models.Metric[models.FlexLegacyMetric]) (id int64, err error) {
	c, err := pg.db.BeginTx(ctx, nil)
	if err != nil {
		return id, err
	}
	id, err = pg.createMetric(ctx, c, metric.Base)
	if err != nil {
		c.Rollback()
		return id, err
	}
	_, err = c.ExecContext(ctx, sqlFlexLegacyMetricsCreate,
		id,
		metric.Protocol.OID,
		metric.Protocol.Port,
		metric.Protocol.PortType,
	)
	if err != nil {
		c.Rollback()
		return id, err
	}
	return id, c.Commit()
}

func (pg *PG) UpdateFlexLegacyMetric(ctx context.Context, metric models.Metric[models.FlexLegacyMetric]) (exists bool, err error) {
	c, err := pg.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	exists, err = pg.updateMetric(ctx, c, metric.Base)
	if err != nil {
		c.Rollback()
		return false, err
	}
	if !exists {
		return false, nil
	}
	t, err := c.ExecContext(ctx, sqlFlexLegacyMetricsUpdate,
		metric.Base.Id,
		metric.Protocol.OID,
		metric.Protocol.Port,
		metric.Protocol.PortType,
	)
	if err != nil {
		c.Rollback()
		return false, err
	}
	rowsAffected, _ := t.RowsAffected()
	return rowsAffected != 0, c.Commit()
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
	rows, err := pg.db.QueryContext(ctx, sqlFlexLegacyMetricsGetProtocol, id)
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
	rows, err := pg.db.QueryContext(ctx, sqlFlexLegacyMetricsGetAsSNMPMetric, id)
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
	rows, err := pg.db.QueryContext(ctx, sqlFlexLegacyMetricsGetByIdsAsSNMPMetric, ids)
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

func (pg *PG) GetFlexLegacyMetricsRequests(ctx context.Context, containerId int32) (metrics []models.FlexLegacyDatalogMetricRequest, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlFlexLegacyMetricsGetMetricsRequests, containerId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	metrics = []models.FlexLegacyDatalogMetricRequest{}
	for rows.Next() {
		var m models.FlexLegacyDatalogMetricRequest
		err = rows.Scan(&m.Id, &m.Type, &m.DataPolicyId, &m.Port, &m.PortType)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, m)
	}
	return metrics, nil
}
