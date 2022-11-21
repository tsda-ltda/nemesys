package pg

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/jackc/pgx/v5"
)

type ContextualMetrics struct {
	*pgx.Conn
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

// CtxMetricsGetResponse is the response for Get handler.
type CtxMetricsGetResponse struct {
	// Exists is the contextual metric existence.
	Exists bool
	// ContextualMetric is the contextual metric.
	ContextualMetric models.ContextualMetric
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
		LEFT JOIN contexts c ON c.ident = $1 WHERE cm.ident = $2;`
	sqlCtxMetricsCreate = `INSERT INTO contextual_metrics
		(ctx_id, metric_id, ident, name, descr) VALUES ($1, $2, $3, $4, $5) RETURNING id;`
	sqlCtxMetricsUpdate = `UPDATE contextual_metrics SET 
		(ident, name, descr) = ($1, $2, $3) WHERE id = $4;`
	sqlCtxMetricsDelete = `DELETE FROM contextual_metrics WHERE id = $1;`
	sqlCtxMetricsMGet   = `SELECT id, metric_id, ident, name, descr FROM contextual_metrics WHERE ctx_id = $1 LIMIT $2 OFFSET $3;`
	sqlCtxMetricsGet    = `SELECT ctx_id, metric_id, ident, name, descr FROM contextual_metrics WHERE id = $1;`
)

// GetIdsByIdent returns a team id, ctx id, metric id using their idents.
func (c *ContextualMetrics) GetIdsByIdent(ctx context.Context, metricIdent string, ctxIdent string, teamIdent string) (r ContextualMetricsGetIdsByIdentResponse, err error) {
	rows, err := c.Query(ctx, sqlCtxMetricsGetIdsByIdent, teamIdent, ctxIdent, metricIdent)
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

// Get returns a contextual metric by id.
func (c *ContextualMetrics) Get(ctx context.Context, id int64) (r CtxMetricsGetResponse, err error) {
	rows, err := c.Query(ctx, sqlCtxMetricsGet, id)
	if err != nil {
		return r, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(
			&r.ContextualMetric.ContextId,
			&r.ContextualMetric.MetricId,
			&r.ContextualMetric.Ident,
			&r.ContextualMetric.Name,
			&r.ContextualMetric.Descr,
		)
		r.ContextualMetric.Id = id
		r.Exists = true
	}
	return r, err
}

// MGet returns all metrics of a context with limit and offset.
func (c *ContextualMetrics) MGet(ctx context.Context, ctxId int32, limit int, offset int) (metrics []models.ContextualMetric, err error) {
	metrics = []models.ContextualMetric{}
	rows, err := c.Query(ctx, sqlCtxMetricsMGet, ctxId, limit, offset)
	if err != nil {
		return metrics, err
	}
	defer rows.Close()
	for rows.Next() {
		var m models.ContextualMetric
		err = rows.Scan(
			&m.Id,
			&m.MetricId,
			&m.Ident,
			&m.Name,
			&m.Descr,
		)
		if err != nil {
			return metrics, err
		}
		m.ContextId = ctxId
		metrics = append(metrics, m)
	}
	return metrics, nil
}

// ExistsIdent returns if a contextual metric's ident exists.
func (c *ContextualMetrics) ExistsIdent(ctx context.Context, ident string, ctxId int32, id int64) (exists bool, err error) {
	return exists, c.QueryRow(ctx, sqlCtxMetricsExistsIdent, ident, ctxId, id).Scan(&exists)
}

// GetMetricEnabledAndRequestById returns the metric's information to make a data request and
// if is enabled, querying by id.
func (c *ContextualMetrics) GetMetricEnabledAndRequestById(ctx context.Context, contextualMetricId int64) (r CtxMetricsGetMetricEnabledAndMetricRequestByIdResponse, err error) {
	rows, err := c.Query(ctx, sqlCtxMetricsGetMetricRequestInfo, contextualMetricId)
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

// ExistsContextAndIdent returns if a context, metric and a contexual metric's ident exists.
func (c *ContextualMetrics) ExistsContextMetricAndIdent(ctx context.Context, contextId int32, metricId int64, ident string, contextualMetricId int64) (r CtxMetricsExistsContextMetricAndIdentResponse, err error) {
	rows, err := c.Query(ctx, sqlCtxMetricsExistsContextMetricAndIdent, contextId, metricId, ident, contextualMetricId)
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

// Create creates a contextual metric, returning it's id.
func (c *ContextualMetrics) Create(ctx context.Context, m models.ContextualMetric) (id int64, err error) {
	return id, c.QueryRow(ctx, sqlCtxMetricsCreate,
		m.ContextId,
		m.MetricId,
		m.Ident,
		m.Name,
		m.Descr,
	).Scan(&id)
}

// Update updates a contextual metric if exists.
func (c *ContextualMetrics) Update(ctx context.Context, m models.ContextualMetric) (exists bool, err error) {
	t, err := c.Exec(ctx, sqlCtxMetricsUpdate,
		m.Ident,
		m.Name,
		m.Descr,
		m.Id,
	)
	return t.RowsAffected() != 0, err
}

// Delete deletes a contextual metric by id.
func (c *ContextualMetrics) Delete(ctx context.Context, id int64) (exists bool, err error) {
	t, err := c.Exec(ctx, sqlCtxMetricsDelete, id)
	return t.RowsAffected() != 0, err
}
