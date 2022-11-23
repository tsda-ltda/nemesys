package pg

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/models"
)

type SNMPv2cMetricGetResponse struct {
	// Exstis is the SNMP metric existence.
	Exists bool
	// Metric is the metric.
	Metric models.Metric[models.SNMPMetric]
}

type SNMPv2cMetricGetProtocolResponse struct {
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

func (pg *PG) CreateSNMPv2cMetric(ctx context.Context, m models.Metric[models.SNMPMetric]) error {
	c, err := pg.db.BeginTx(ctx, nil)
	if err != nil {
		c.Rollback()
		return err
	}
	id, err := pg.createMetric(ctx, c, m.Base)
	if err != nil {
		c.Rollback()
		return err
	}
	_, err = pg.db.ExecContext(ctx, sqlSNMPMetricsCreate, m.Protocol.OID, id)
	if err != nil {
		c.Rollback()
		return err
	}
	return c.Commit()
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
	t, err := pg.db.ExecContext(ctx, sqlSNMPMetricsUpdate, m.Protocol.OID, m.Protocol.Id, m.Protocol.Id)
	if err != nil {
		c.Rollback()
		return false, err
	}
	rowsAffected, _ := t.RowsAffected()
	return rowsAffected != 0, c.Commit()
}

func (pg *PG) GetSNMPv2cMetric(ctx context.Context, id int64) (r SNMPv2cMetricGetResponse, err error) {
	baseR, err := pg.GetMetric(ctx, id)
	if err != nil {
		return r, err
	}
	protocolR, err := pg.GetSNMPv2cMetricProtocol(ctx, id)
	if err != nil {
		return r, err
	}
	r.Exists = baseR.Exists
	r.Metric.Base = baseR.Metric
	r.Metric.Protocol = protocolR.Metric
	return r, err
}

func (pg *PG) GetSNMPv2cMetricProtocol(ctx context.Context, id int64) (r SNMPv2cMetricGetProtocolResponse, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlSNMPMetricsGet, id)
	if err != nil {
		return r, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&r.Metric.OID)
		if err != nil {
			return r, err
		}
		r.Metric.Id = id
		r.Exists = true
	}
	return r, nil
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
