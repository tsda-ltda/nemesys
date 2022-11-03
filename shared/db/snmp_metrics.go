package db

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/jackc/pgx/v5"
)

type SNMPMetrics struct {
	*pgx.Conn
}

const (
	sqlSNMPMetricsGet    = `SELECT oid FROM snmp_metrics WHERE metric_id = $1;`
	sqlSNMPMetricsCreate = `INSERT INTO snmp_metrics (oid, metric_id) VALUES ($1, $2);`
	sqlSNMPMetricsUpdate = `UPDATE snmp_metrics SET (oid, metric_id) = ($1, $2) WHERE metric_id = $3;`
)

// Create creates a SNMP metric. Returns an error if fail to create.
func (c *SNMPMetrics) Create(ctx context.Context, m models.SNMPMetric) error {
	_, err := c.Exec(ctx, sqlSNMPMetricsCreate, m.OID, m.MetricId)
	return err
}

// Update updates a SNMP metric if exists. Returns an error if fail to create.
func (c *SNMPMetrics) Update(ctx context.Context, m models.SNMPMetric) (e bool, err error) {
	t, err := c.Exec(ctx, sqlSNMPMetricsUpdate, m.OID, m.MetricId, m.MetricId)
	return t.RowsAffected() != 0, err
}

// Get returns a SNMP metric if exists. Returns an error if fails to get.
func (c *SNMPMetrics) Get(ctx context.Context, metricId int64) (e bool, m models.SNMPMetric, err error) {
	rows, err := c.Query(ctx, sqlSNMPMetricsGet, metricId)
	if err != nil {
		return false, m, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&m.OID)
		if err != nil {
			return false, m, err
		}
		m.MetricId = metricId
		e = true
	}
	return e, m, nil
}
