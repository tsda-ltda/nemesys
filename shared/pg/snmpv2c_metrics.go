package pg

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/models"
)

const (
	sqlSNMPMetricsGet      = `SELECT oid FROM snmpv2c_metrics WHERE metric_id = $1;`
	sqlSNMPMetricsGetByIds = `SELECT metric_id, oid FROM snmpv2c_metrics WHERE metric_id = ANY ($1);`
	sqlSNMPMetricsCreate   = `INSERT INTO snmpv2c_metrics (oid, metric_id) VALUES ($1, $2);`
	sqlSNMPMetricsUpdate   = `UPDATE snmpv2c_metrics SET (oid, metric_id) = ($1, $2) WHERE metric_id = $3;`
)

func (pg *PG) CreateSNMPv2cMetric(ctx context.Context, m models.Metric[models.SNMPMetric]) (id int64, err error) {
	c, err := pg.db.BeginTx(ctx, nil)
	if err != nil {
		c.Rollback()
		return id, err
	}
	id, err = pg.createMetric(ctx, c, m.Base)
	if err != nil {
		c.Rollback()
		return id, err
	}
	_, err = c.ExecContext(ctx, sqlSNMPMetricsCreate, m.Protocol.OID, id)
	if err != nil {
		c.Rollback()
		return id, err
	}
	return id, c.Commit()
}

func (pg *PG) UpdateSNMPv2cMetric(ctx context.Context, m models.Metric[models.SNMPMetric]) (exists bool, err error) {
	c, err := pg.db.BeginTx(ctx, nil)
	if err != nil {
		c.Rollback()
		return false, err
	}
	exists, err = pg.updateMetric(ctx, c, m.Base)
	if err != nil {
		c.Rollback()
		return false, err
	}
	if !exists {
		return false, nil
	}
	t, err := c.ExecContext(ctx, sqlSNMPMetricsUpdate, m.Protocol.OID, m.Protocol.Id, m.Protocol.Id)
	if err != nil {
		c.Rollback()
		return false, err
	}
	rowsAffected, _ := t.RowsAffected()
	return rowsAffected != 0, c.Commit()
}

func (pg *PG) GetSNMPv2cMetric(ctx context.Context, id int64) (exists bool, metric models.Metric[models.SNMPMetric], err error) {
	exists, base, err := pg.GetMetric(ctx, id)
	if err != nil {
		return false, metric, err
	}
	if !exists {
		return false, metric, nil
	}
	exists, protocol, err := pg.GetSNMPv2cMetricProtocol(ctx, id)
	if err != nil {
		return false, metric, err
	}
	metric.Base = base
	metric.Protocol = protocol
	return exists, metric, err
}

func (pg *PG) GetSNMPv2cMetricProtocol(ctx context.Context, id int64) (exists bool, metric models.SNMPMetric, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlSNMPMetricsGet, id)
	if err != nil {
		return false, metric, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&metric.OID)
		if err != nil {
			return exists, metric, err
		}
		metric.Id = id
		exists = true
	}
	return exists, metric, nil
}

func (pg *PG) GetSNMPv2cMetricsByIds(ctx context.Context, ids []int64) (metrics []models.SNMPMetric, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlSNMPMetricsGetByIds, ids)
	if err != nil {
		return nil, err
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
