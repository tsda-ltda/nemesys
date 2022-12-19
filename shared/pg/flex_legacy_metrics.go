package pg

import (
	"context"
	"database/sql"

	"github.com/fernandotsda/nemesys/shared/models"
)

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
	sqlFlexLegacyMetricsGetIdByPortPortType = `SELECT m.id FROM metrics m LEFT JOIN flex_legacy_metrics fm ON m.id = fm.metric_id 
		WHERE m.container_id = $1 AND fm.port = $2 AND fm.port_type = $3`
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
	exists, base, err := pg.GetMetric(ctx, id)
	if err != nil {
		return exists, metric, err
	}
	if !exists {
		return false, metric, nil
	}
	exists, protocol, err := pg.GetFlexLegacyMetricProtocol(ctx, id)
	if err != nil {
		return exists, metric, err
	}
	metric.Base = base
	metric.Protocol = protocol
	return exists, metric, nil
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
