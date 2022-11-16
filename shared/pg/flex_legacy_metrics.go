package pg

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/jackc/pgx/v5"
)

type FlexLegacyMetrics struct {
	*pgx.Conn
}

// FlexLegacyMetricsGetResponse is the response for Get handler.
type FlexLegacyMetricsGetResponse struct {
	// Exists is the metric existence.
	Exists bool
	// Metric is the metric.
	Metric models.FlexLegacyMetric
}

// FlexLegacyMetricsGetAsSNMPMetricResponse is the response for GetAsSNMPMetric handler.
type FlexLegacyMetricsGetAsSNMPMetricResponse struct {
	// Exists is the metric existence.
	Exists bool
	// Metric is the metric.
	Metric models.SNMPMetric
}

// FlexLegacyMetricsGetByIdsAsSNMPMetricResponse is the response for GetByIdsAsSNMPMetric handler.
type FlexLegacyMetricsGetByIdsAsSNMPMetricResponse struct {
	// Exists is the metric existence.
	Exists bool
	// Metric is the metric.
	Metric models.SNMPMetric
}

const (
	sqlFlexLegacyMetricsCreate               = `INSERT INTO flex_legacy_metrics (metric_id, oid, port, port_type) VALUES($1, $2, $3, $4);`
	sqlFlexLegacyMetricsUpdate               = `UPDATE flex_legacy_metrics SET (oid, port, port_type) = ($2, $3, $4) WHERE metric_id = $1;`
	sqlFlexLegacyMetricsGet                  = `SELECT oid, port, port_type FROM flex_legacy_metrics WHERE metric_id = $1;`
	sqlFlexLegacyMetricsGetAsSNMPMetric      = `SELECT oid FROM flex_legacy_metrics WHERE metric_id = $1;`
	sqlFlexLegacyMetricsGetByIdsAsSNMPMetric = `SELECT metric_id, oid FROM flex_legacy_metrics WHERE metric_id = ANY ($1);`
)

// Create creates a flex legacy metric.
func (c *FlexLegacyMetrics) Create(ctx context.Context, metric models.FlexLegacyMetric) (err error) {
	_, err = c.Exec(ctx, sqlFlexLegacyMetricsCreate,
		metric.Id,
		metric.OID,
		metric.Port,
		metric.PortType,
	)
	return err
}

// Update updates a flex legacy metric.
func (c *FlexLegacyMetrics) Update(ctx context.Context, metric models.FlexLegacyMetric) (exists bool, err error) {
	t, err := c.Exec(ctx, sqlFlexLegacyMetricsUpdate,
		metric.Id,
		metric.OID,
		metric.Port,
		metric.PortType,
	)
	return t.RowsAffected() != 0, err
}

// Get returns a flex legacy metric.
func (c *FlexLegacyMetrics) Get(ctx context.Context, id int64) (r FlexLegacyMetricsGetResponse, err error) {
	rows, err := c.Query(ctx, sqlFlexLegacyMetricsGet, id)
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

// GetAsSNMPMetric returns a flex metric as a SNMP metric.
func (c *FlexLegacyMetrics) GetAsSNMPMetric(ctx context.Context, id int64) (r FlexLegacyMetricsGetAsSNMPMetricResponse, err error) {
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

// GetByIdsAsSNMPMetric returns all flex metric as a SNMP metric.
func (c *FlexLegacyMetrics) GetByIdsAsSNMPMetric(ctx context.Context, ids []int64) (metrics []models.SNMPMetric, err error) {
	rows, err := c.Query(ctx, sqlFlexLegacyMetricsGetByIdsAsSNMPMetric, ids)
	if err != nil {
		return metrics, err
	}
	defer rows.Close()
	for rows.Next() {
		var m models.SNMPMetric
		err = rows.Scan(
			&m.Id,
			&m.OID,
		)
		if err != nil {
			return metrics, err
		}
		metrics = append(metrics, m)
	}
	return metrics, nil
}
