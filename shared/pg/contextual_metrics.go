package pg

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/models"
)

var ContextualMetricValidOrderByColumns = []string{"name", "descr", "ident"}

type ContextualMetricQueryFilters struct {
	CtxId     int32  `type:"=" column:"ctx_id"`
	Name      string `type:"ilike" column:"name"`
	Descr     string `type:"ilike" column:"descr"`
	Ident     string `type:"ilike" column:"ident"`
	OrderBy   string
	OrderByFn string
	Limit     int
	Offset    int
}

func (f ContextualMetricQueryFilters) GetOrderBy() string {
	return f.OrderBy
}

func (f ContextualMetricQueryFilters) GetOrderByFn() string {
	return f.OrderByFn
}

func (f ContextualMetricQueryFilters) GetLimit() int {
	return f.Limit
}

func (f ContextualMetricQueryFilters) GetOffset() int {
	return f.Offset
}

// ContextualMetricsGetIdsByIdentResponse is the response for GetIdsByIdent handler.
type ContextualMetricsGetIdsByIdentResponse struct {
	// Exists is the contextual metric existence.
	Exists bool
	// ContextualMetricId is the contextual metric id.
	ContextualMetricId int64
	// ContextId is the context id.
	ContextId int32
	// TeamId is the team id.
	TeamId int32
}

// CtxMetricsGetMetricEnabledAndMetricRequestByIdResponse is the response for GetMetricEnabledAndRequestById handler.
type CtxMetricsGetMetricEnabledAndMetricRequestByIdResponse struct {
	// Exists is the contextual metric existence.
	Exists bool
	// Enabled is the metric enabled status.
	Enabled bool
	// MetricRequest is the metric request information.
	MetricRequest models.MetricRequest
}

// CtxMetricsExistsContextMetricAndIdentResponse is the response for ExistsContextMetricAndIdent handler
type CtxMetricsExistsContextMetricAndIdentResponse struct {
	// ContextExists is the context existence.
	ContextExists bool
	// MetricExists is the metric existence.
	MetricExists bool
	// IdentExists is the contextual metric ident existence.
	IdentExists bool
}

const (
	sqlCtxMetricsGetMetricRequestInfo = `SELECT id, enabled, type, container_id, container_type, data_policy_id FROM metrics 
		WHERE id = (SELECT metric_id FROM contextual_metrics WHERE id = $1);`
	sqlCtxMetricsGetIdsByIdent = `WITH 
		tid AS (SELECT id FROM teams WHERE ident = $1),
		cid AS (SELECT id FROM contexts WHERE ident = $2 and team_id = (SELECT * FROM tid))
		SELECT id, (SELECT * FROM cid), (SELECT * FROM tid) FROM contextual_metrics WHERE ident = $3 AND ctx_id = (SELECT * FROM cid);`
	sqlCtxMetricsExistsIdent = `SELECT EXISTS (SELECT 1 FROM contextual_metrics 
		WHERE ident = $1 AND ctx_id = $2 AND id != $3);`
	sqlCtxMetricsExistsContextMetricAndIdent = `SELECT
		EXISTS (SELECT 1 FROM contexts WHERE id = $1),
		EXISTS (SELECT 1 FROM metrics WHERE id = $2), 
		EXISTS (SELECT 1 FROM contextual_metrics WHERE ident=$3 AND id != $4 and ctx_id = $1);`
	sqlCtxMetricsGetIdByIdent = `SELECT cm.id FROM contextual_metrics cm 
		LEFT JOIN contexts c ON pg.pool.ident = $1 WHERE cm.ident = $2;`
	sqlCtxMetricsCreate = `INSERT INTO contextual_metrics
		(ctx_id, metric_id, ident, name, descr) VALUES ($1, $2, $3, $4, $5) RETURNING id;`
	sqlCtxMetricsUpdate = `UPDATE contextual_metrics SET 
		(ident, name, descr) = ($1, $2, $3) WHERE id = $4;`
	sqlCtxMetricsDelete = `DELETE FROM contextual_metrics WHERE id = $1;`
	sqlCtxMetricsGet    = `SELECT ctx_id, metric_id, ident, name, descr FROM contextual_metrics WHERE id = $1;`

	customSqlCtxMetricsMGet = `SELECT id, metric_id, ident, name, descr FROM contextual_metrics`
)

func (pg *PG) GetContextualMetricTreeId(ctx context.Context, metricIdent string, ctxIdent string, teamIdent string) (r ContextualMetricsGetIdsByIdentResponse, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlCtxMetricsGetIdsByIdent, teamIdent, ctxIdent, metricIdent)
	if err != nil {
		return r, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(
			&r.ContextualMetricId,
			&r.ContextId,
			&r.TeamId,
		)
		if err != nil {
			return r, err
		}
		r.Exists = true
	}
	return r, nil
}

func (pg *PG) GetContextualMetric(ctx context.Context, id int64) (exists bool, metric models.ContextualMetric, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlCtxMetricsGet, id)
	if err != nil {
		return false, metric, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(
			&metric.ContextId,
			&metric.MetricId,
			&metric.Ident,
			&metric.Name,
			&metric.Descr,
		)
		metric.Id = id
		exists = true
	}
	return exists, metric, err
}

func (pg *PG) GetContextualMetrics(ctx context.Context, filters ContextualMetricQueryFilters) (metrics []models.ContextualMetric, err error) {
	sql, params, err := applyFilters(filters, customSqlCtxMetricsMGet, ContextualMetricValidOrderByColumns)
	if err != nil {
		return nil, err
	}
	rows, err := pg.db.QueryContext(ctx, sql, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	metrics = make([]models.ContextualMetric, 0, filters.Limit)
	var m models.ContextualMetric
	for rows.Next() {
		err = rows.Scan(
			&m.Id,
			&m.MetricId,
			&m.Ident,
			&m.Name,
			&m.Descr,
		)
		if err != nil {
			return nil, err
		}
		m.ContextId = filters.CtxId
		metrics = append(metrics, m)
	}
	return metrics, nil
}

func (pg *PG) ContextualMetricIdentExists(ctx context.Context, ident string, ctxId int32, id int64) (exists bool, err error) {
	return exists, pg.db.QueryRowContext(ctx, sqlCtxMetricsExistsIdent, ident, ctxId, id).Scan(&exists)
}

func (pg *PG) GetMetricRequestByContextualMetric(ctx context.Context, contextualMetricId int64) (r CtxMetricsGetMetricEnabledAndMetricRequestByIdResponse, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlCtxMetricsGetMetricRequestInfo, contextualMetricId)
	if err != nil {
		return r, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(
			&r.MetricRequest.MetricId,
			&r.Enabled,
			&r.MetricRequest.MetricType,
			&r.MetricRequest.ContainerId,
			&r.MetricRequest.ContainerType,
			&r.MetricRequest.DataPolicyId,
		)
		if err != nil {
			return r, err
		}
		r.Exists = true
	}
	return r, nil
}

func (pg *PG) ContextMetricAndContexualMetricIdentExists(ctx context.Context, contextId int32, metricId int64, ident string, contextualMetricId int64) (r CtxMetricsExistsContextMetricAndIdentResponse, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlCtxMetricsExistsContextMetricAndIdent, contextId, metricId, ident, contextualMetricId)
	if err != nil {
		return r, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(
			&r.ContextExists,
			&r.MetricExists,
			&r.IdentExists,
		)
		if err != nil {
			return r, err
		}
	}
	return r, nil
}

func (pg *PG) CreateContextualMetric(ctx context.Context, m models.ContextualMetric) (id int64, err error) {
	return id, pg.db.QueryRowContext(ctx, sqlCtxMetricsCreate,
		m.ContextId,
		m.MetricId,
		m.Ident,
		m.Name,
		m.Descr,
	).Scan(&id)
}

func (pg *PG) UpdateContextualMetric(ctx context.Context, m models.ContextualMetric) (exists bool, err error) {
	t, err := pg.db.ExecContext(ctx, sqlCtxMetricsUpdate,
		m.Ident,
		m.Name,
		m.Descr,
		m.Id,
	)
	if err != nil {
		return false, err
	}
	rowsAffected, _ := t.RowsAffected()
	return rowsAffected != 0, err
}

func (pg *PG) DeleteContextualMetric(ctx context.Context, id int64) (exists bool, err error) {
	t, err := pg.db.ExecContext(ctx, sqlCtxMetricsDelete, id)
	if err != nil {
		return false, err
	}
	rowsAffected, _ := t.RowsAffected()
	return rowsAffected != 0, err
}
