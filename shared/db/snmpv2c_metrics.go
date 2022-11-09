package db

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/jackc/pgx/v5"
)

type SNMPv2cMetrics struct {
	*pgx.Conn
}

const (
	sqlSNMPMetricsGet      = `SELECT oid FROM snmpv2c_metrics WHERE metric_id = $1;`
	sqlSNMPMetricsGetByIds = `SELECT metric_id, oid FROM snmpv2c_metrics WHERE metric_id = ANY ($1);`
	sqlSNMPMetricsCreate   = `INSERT INTO snmpv2c_metrics (oid, metric_id) VALUES ($1, $2);`
	sqlSNMPMetricsUpdate   = `UPDATE snmpv2c_metrics SET (oid, metric_id) = ($1, $2) WHERE metric_id = $3;`
)

// Create creates a SNMP metric. Returns an error if fail to create.
func (c *SNMPv2cMetrics) Create(ctx context.Context, m models.SNMPMetric) error {
	_, err := c.Exec(ctx, sqlSNMPMetricsCreate, m.OID, m.Id)
	return err
}

// Update updates a SNMP metric if exists. Returns an error if fail to create.
func (c *SNMPv2cMetrics) Update(ctx context.Context, m models.SNMPMetric) (e bool, err error) {
	t, err := c.Exec(ctx, sqlSNMPMetricsUpdate, m.OID, m.Id, m.Id)
	return t.RowsAffected() != 0, err
}

// Get returns a SNMP metric if exists. Returns an error if fails to get.
func (c *SNMPv2cMetrics) Get(ctx context.Context, metricId int64) (e bool, m models.SNMPMetric, err error) {
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
		m.Id = metricId
		e = true
	}
	return e, m, nil
}

// Get returns multiple SNMP metric by an array of ids. Returns an error if fails to get.
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
