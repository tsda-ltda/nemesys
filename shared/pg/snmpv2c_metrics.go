package pg

import (
	"context"
	"database/sql"

	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/types"
)

var SNMPv2cMetricValidOrderByColumns = []string{"name", "descr"}

type SNMPv2cMetricQueryFilters struct {
	ContainerType types.ContainerType `type:"=" column:"container_type"`
	ContainerId   int32               `type:"=" column:"container_id"`
	Name          string              `type:"ilike" column:"name"`
	Descr         string              `type:"ilike" column:"descr"`
	Enabled       *bool               `type:"=" column:"enabled"`
	DataPolicyId  int16               `type:"=" column:"data_policy_id"`
	OrderBy       string
	OrderByFn     string
	Limit         int
	Offset        int
}

func (f SNMPv2cMetricQueryFilters) GetOrderBy() string {
	return f.OrderBy
}

func (f SNMPv2cMetricQueryFilters) GetOrderByFn() string {
	return f.OrderByFn
}

func (f SNMPv2cMetricQueryFilters) GetLimit() int {
	return f.Limit
}

func (f SNMPv2cMetricQueryFilters) GetOffset() int {
	return f.Offset
}

const (
	sqlSNMPv2cMetricsGet = `SELECT 
		b.container_id, b.name, b.descr, b.enabled, b.data_policy_id, 
		b.rts_pulling_times, b.rts_data_cache_duration, b.dhs_enabled, b.dhs_interval, b.type, b.ev_expression 
		p.oid FROM metrics b FULL JOIN snmpv2c_metrics p ON p.metric_id = b.id WHERE id = $1;`
	sqlSNMPv2cMetricsGetByIds                   = `SELECT metric_id, oid FROM snmpv2c_metrics WHERE metric_id = ANY ($1);`
	sqlSNMPv2cMetricsCreate                     = `INSERT INTO snmpv2c_metrics (oid, metric_id) VALUES ($1, $2);`
	sqlSNMPv2cMetricsUpdate                     = `UPDATE snmpv2c_metrics SET (oid, metric_id) = ($1, $2) WHERE metric_id = $3;`
	customSqlBasicMetricsMGetSNMPv2cMetricsMGet = `SELECT 
	b.id, b.name, b.descr, b.enabled, b.data_policy_id, 
	b.rts_pulling_times, b.rts_data_cache_duration, b.dhs_enabled, b.dhs_interval, b.type, b.ev_expression 
	p.oid FROM metrics b FULL JOIN snmpv2c_metrics p ON p.metric_id = b.id %s LIMIT $1 OFFSET $2`
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
	_, err = c.ExecContext(ctx, sqlSNMPv2cMetricsCreate, m.Protocol.OID, id)
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
	t, err := c.ExecContext(ctx, sqlSNMPv2cMetricsUpdate, m.Protocol.OID, m.Protocol.Id, m.Protocol.Id)
	if err != nil {
		c.Rollback()
		return false, err
	}
	rowsAffected, _ := t.RowsAffected()
	return rowsAffected != 0, c.Commit()
}

func (pg *PG) GetSNMPv2cMetric(ctx context.Context, id int64) (exists bool, metric models.Metric[models.SNMPMetric], err error) {
	err = pg.db.QueryRowContext(ctx, sqlFlexLegacyMetricsGet, id).Scan(
		&metric.Base.ContainerId,
		&metric.Base.Name,
		&metric.Base.Descr,
		&metric.Base.Enabled,
		&metric.Base.DataPolicyId,
		&metric.Base.RTSPullingTimes,
		&metric.Base.RTSCacheDuration,
		&metric.Base.DHSEnabled,
		&metric.Base.DHSInterval,
		&metric.Base.Type,
		&metric.Base.EvaluableExpression,
		&metric.Protocol.OID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, metric, nil
		}
		return false, metric, err
	}
	metric.Base.Id = id
	metric.Base.ContainerType = types.CTSNMPv2c
	metric.Protocol.Id = id
	return true, metric, nil
}

func (pg *PG) GetSNMPv2cMetrics(ctx context.Context, filters SNMPv2cMetricQueryFilters) (metrics []models.Metric[models.SNMPMetric], err error) {
	filters.ContainerType = types.CTSNMPv2c
	sql, params, err := applyFilters(filters, customSqlFlexLegacyMetricsMGet, SNMPv2cMetricValidOrderByColumns)
	if err != nil {
		return nil, err
	}
	rows, err := pg.db.QueryContext(ctx, sql, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	metrics = make([]models.Metric[models.SNMPMetric], 0, filters.Limit)
	var metric models.Metric[models.SNMPMetric]
	metric.Base.ContainerId = filters.ContainerId
	metric.Base.ContainerType = filters.ContainerType
	for rows.Next() {
		err = rows.Scan(
			&metric.Base.Id,
			&metric.Base.Name,
			&metric.Base.Descr,
			&metric.Base.Enabled,
			&metric.Base.DataPolicyId,
			&metric.Base.RTSPullingTimes,
			&metric.Base.RTSCacheDuration,
			&metric.Base.DHSEnabled,
			&metric.Base.DHSInterval,
			&metric.Base.Type,
			&metric.Base.EvaluableExpression,
			&metric.Protocol.OID,
		)
		if err != nil {
			return nil, err
		}
		metric.Protocol.Id = metric.Base.Id

	}
	return metrics, nil
}

func (pg *PG) GetSNMPv2cMetricProtocol(ctx context.Context, id int64) (exists bool, metric models.SNMPMetric, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlSNMPv2cMetricsGet, id)
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
	rows, err := pg.db.QueryContext(ctx, sqlSNMPv2cMetricsGetByIds, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	metrics = []models.SNMPMetric{}
	var m models.SNMPMetric
	for rows.Next() {
		err = rows.Scan(&m.Id, &m.OID)
		if err != nil {
			return metrics, err
		}
		metrics = append(metrics, m)
	}
	return metrics, nil
}
