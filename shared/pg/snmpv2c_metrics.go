package pg

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/jackc/pgx/v5"
)

type SNMPv2cMetrics struct {
	*pgx.Conn
}

// SNMPv2cMetricsGetResponse is the response for Get handler.
type SNMPv2cMetricsGetResponse struct {
	// Exstis is the SNMP metric existence.
	Exists bool
	// Metric is the metric.
	Metric models.SNMPMetric
}

const (
	sqlSNMPMetricsGet      = `SELECT oid FROM snmpv2c_metrics WHERE metric_id = $1;`
	sqlSNMPMetricsGetByIds = `SELECT metric_id, oid FROM snmpv2c_metrics WHERE metric_id = ANY ($1);`
	sqlSNMPMetricsCreate   = `INSERT INTO snmpv2c_metrics (oid, metric_id) VALUES ($1, $2);`
	sqlSNMPMetricsUpdate   = `UPDATE snmpv2c_metrics SET (oid, metric_id) = ($1, $2) WHERE metric_id = $3;`
)

// Create creates a SNMP metric.
func (c *SNMPv2cMetrics) Create(ctx context.Context, m models.SNMPMetric) error {
	_, err := c.Exec(ctx, sqlSNMPMetricsCreate, m.OID, m.Id)
	return err
}

// Update updates a SNMP metric if exists.
func (c *SNMPv2cMetrics) Update(ctx context.Context, m models.SNMPMetric) (exists bool, err error) {
	t, err := c.Exec(ctx, sqlSNMPMetricsUpdate, m.OID, m.Id, m.Id)
	return t.RowsAffected() != 0, err
}

// Get returns a SNMP metric if exists.
func (c *SNMPv2cMetrics) Get(ctx context.Context, metricId int64) (r SNMPv2cMetricsGetResponse, err error) {
	rows, err := c.Query(ctx, sqlSNMPMetricsGet, metricId)
	if err != nil {
		return r, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&r.Metric.OID)
		if err != nil {
			return r, err
		}
		r.Metric.Id = metricId
		r.Exists = true
	}
	return r, nil
}

// Get returns multiple SNMP metric by an array of ids.
func (c *SNMPv2cMetrics) GetByIds(ctx context.Context, ids []int64) (metrics []models.SNMPMetric, err error) {
	rows, err := c.Query(ctx, sqlSNMPMetricsGetByIds, ids)
	if err != nil {
		return metrics, err
	}
	defer rows.Close()
	metrics = []models.SNMPMetric{}
	for rows.Next() {
		var m models.SNMPMetric
		err = rows.Scan(&m.Id, &m.OID)
		if err != nil {
			return metrics, err
		}
		metrics = append(metrics, m)
	}
	return metrics, nil
}
