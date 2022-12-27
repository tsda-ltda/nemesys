package pg

import (
	"context"
	"database/sql"

	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/types"
)

var FlexLegacyMetricValidOrderByColumns = []string{"name", "descr", "port", "port_type"}

type FlexLegacyMetricQueryFilters struct {
	ContainerType types.ContainerType `type:"=" column:"container_type"`
	ContainerId   int32               `type:"=" column:"container_id"`
	Name          string              `type:"ilike" column:"name"`
	Descr         string              `type:"ilike" column:"descr"`
	Enabled       *bool               `type:"=" column:"enabled"`
	Port          int16               `type:"=" column:"port"`
	PortType      int16               `type:"=" column:"port_type"`
	OrderBy       string
	OrderByFn     string
}

func (f FlexLegacyMetricQueryFilters) GetOrderBy() string {
	return f.OrderBy
}

func (f FlexLegacyMetricQueryFilters) GetOrderByFn() string {
	return f.OrderByFn
}

const (
	sqlFlexLegacyMetricsCreate = `INSERT INTO flex_legacy_metrics (metric_id, oid, port, port_type) VALUES($1, $2, $3, $4);`
	sqlFlexLegacyMetricsUpdate = `UPDATE flex_legacy_metrics SET (oid, port, port_type) = ($2, $3, $4) WHERE metric_id = $1;`
	sqlFlexLegacyMetricsGet    = `SELECT 
		b.container_id, b.name, b.descr, b.enabled, b.data_policy_id, 
		b.rts_pulling_times, b.rts_data_cache_duration, b.dhs_enabled, b.dhs_interval, b.type, b.ev_expression, 
		p.oid, p.port, p.port_type FROM metrics b FULL JOIN flex_legacy_metrics p ON p.metric_id = b.id WHERE id = $1;`
	sqlFlexLegacyMetricsGetProtocol          = `SELECT oid, port, port_type FROM flex_legacy_metrics WHERE metric_id = $1;`
	sqlFlexLegacyMetricsGetAsSNMPMetric      = `SELECT oid FROM flex_legacy_metrics WHERE metric_id = $1;`
	sqlFlexLegacyMetricsGetByIdsAsSNMPMetric = `SELECT metric_id, oid FROM flex_legacy_metrics WHERE metric_id = ANY ($1);`
	sqlFlexLegacyMetricsGetMetricsRequests   = `SELECT
		m.id, m.type, m.data_policy_id, f.port, f.port_type 
		FROM metrics m FULL JOIN flex_legacy_metrics f ON m.id = f.metric_id
		WHERE m.enabled = true AND m.container_id = $1 AND m.dhs_enabled = true;`
	sqlFlexLegacyMetricsGetIdByPortPortType = `SELECT m.id FROM metrics m LEFT JOIN flex_legacy_metrics fm ON m.id = fm.metric_id 
		WHERE m.container_id = $1 AND fm.port = $2 AND fm.port_type = $3`
	customSqlFlexLegacyMetricsMGet = `SELECT 
		b.id, b.container_id, b.name, b.descr, b.enabled, b.data_policy_id, 
		b.rts_pulling_times, b.rts_data_cache_duration, b.dhs_enabled, b.dhs_interval, b.type, b.ev_expression, 
		p.oid, p.port, p.port_type FROM metrics b FULL JOIN flex_legacy_metrics p ON p.metric_id = b.id %s LIMIT $1 AND OFFSET $2`
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

func (pg *PG) GetFlexLegacyMetric(ctx context.Context, id int64) (exists bool, metric models.Metric[models.FlexLegacyMetric], err error) {
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
		&metric.Protocol.Port,
		&metric.Protocol.PortType,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, metric, nil
		}
		return false, metric, err
	}
	metric.Base.Id = id
	metric.Base.ContainerType = types.CTFlexLegacy
	metric.Protocol.Id = id
	return true, metric, nil
}

func (pg *PG) GetFlexLegacyMetrics(ctx context.Context, filters FlexLegacyMetricQueryFilters, limit int, offset int) (metrics []models.Metric[models.FlexLegacyMetric], err error) {
	filters.ContainerType = types.CTFlexLegacy
	sql, err := applyFilters(filters, customSqlFlexLegacyMetricsMGet, FlexLegacyMetricValidOrderByColumns)
	if err != nil {
		return nil, err
	}
	rows, err := pg.db.QueryContext(ctx, sql, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	metrics = make([]models.Metric[models.FlexLegacyMetric], 0, limit)
	var metric models.Metric[models.FlexLegacyMetric]
	metric.Base.ContainerId = filters.ContainerId
	metric.Base.ContainerType = types.CTFlexLegacy
	for rows.Next() {
		err = rows.Scan(
			&metric.Base.Id,
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
			&metric.Protocol.Port,
			&metric.Protocol.PortType,
		)
		if err != nil {
			return nil, err
		}
		metric.Protocol.Id = metric.Base.Id
		metrics = append(metrics, metric)
	}
	return metrics, nil
}
func (pg *PG) GetFlexLegacyMetricProtocol(ctx context.Context, id int64) (exists bool, metric models.FlexLegacyMetric, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlFlexLegacyMetricsGetProtocol, id)
	if err != nil {
		return false, metric, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(
			&metric.OID,
			&metric.Port,
			&metric.PortType,
		)
		if err != nil {
			return false, metric, err
		}
		metric.Id = id
		exists = true
	}
	return exists, metric, nil
}

func (pg *PG) DeleteFlexLegacyMetric(ctx context.Context, id int64) (exists bool, err error) {
	return pg.DeleteMetric(ctx, id)
}

func (pg *PG) GetFlexLegacyMetricAsSNMPMetric(ctx context.Context, id int64) (exists bool, metric models.SNMPMetric, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlFlexLegacyMetricsGetAsSNMPMetric, id)
	if err != nil {
		return false, metric, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(
			&metric.OID,
		)
		if err != nil {
			return false, metric, err
		}
		metric.Id = id
		exists = true
	}
	return exists, metric, nil
}

func (pg *PG) FlexLegacyMetricsByIdsAsSNMPMetric(ctx context.Context, ids []int64) (metrics []models.SNMPMetric, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlFlexLegacyMetricsGetByIdsAsSNMPMetric, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var m models.SNMPMetric
	for rows.Next() {
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
	var m models.FlexLegacyDatalogMetricRequest
	for rows.Next() {
		err = rows.Scan(&m.Id, &m.Type, &m.DataPolicyId, &m.Port, &m.PortType)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, m)
	}
	return metrics, nil
}

func (pg *PG) GetFlexLegacyMetricByPortPortType(ctx context.Context, containerId int32, port int16, portType int16) (exists bool, id int64, err error) {
	err = pg.db.QueryRowContext(ctx, sqlFlexLegacyMetricsGetIdByPortPortType, containerId, port, portType).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, id, nil
		}
		return false, id, err
	}
	return true, id, nil
}
