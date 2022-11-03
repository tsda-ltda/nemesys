package db

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/jackc/pgx/v5"
)

type ContextualMetrics struct {
	*pgx.Conn
}

const (
	sqlGetIdContainerIdTeamId = `WITH 
		tid  AS (SELECT id FROM teams WHERE ident = $1),
		cid  AS (SELECT id FROM contexts WHERE ident = $2 and team_id = (SELECT * FROM tid)),
		mid AS (SELECT metric_id FROM contextual_metrics WHERE ident = $3 AND ctx_id = (SELECT * FROM cid))
		SELECT (SELECT * FROM mid), type, container_id, container_type FROM metrics WHERE id = (SELECT * FROM mid);`
	sqlCtxMetricsExistsIdent = `SELECT EXISTS (SELECT 1 FROM contextual_metrics 
		WHERE ident = $1 AND ctx_id = $2 AND id != $3);`
	sqlCtxMetricsExistsContextMetricAndIdent = `SELECT
		EXISTS (SELECT 1 FROM contexts WHERE id = $1),
		EXISTS (SELECT 1 FROM metrics WHERE id = $2), 
		EXISTS (SELECT 1 FROM contextual_metrics WHERE ident=$3 AND id != $4 and ctx_id = $1);`
	sqlCtxMetricsGetIdByIdent = `SELECT cm.id FROM contextual_metrics cm 
		LEFT JOIN contexts c ON c.ident = $1 WHERE cm.ident = $2;`
	sqlCtxMetricsCreate = `INSERT INTO contextual_metrics
		(ctx_id, metric_id, ident, name, descr) VALUES ($1, $2, $3, $4, $5);`
	sqlCtxMetricsUpdate = `UPDATE contextual_metrics SET 
		(ident, name, descr) = ($1, $2, $3) WHERE id = $4;`
	sqlCtxMetricsDelete = `DELETE FROM contextual_metrics WHERE id = $1;`
	sqlCtxMetricsMGet   = `SELECT id, metric_id, ident, name, descr FROM contextual_metrics WHERE ctx_id = $1 LIMIT $2 OFFSET $3;`
	sqlCtxMetricsGet    = `SELECT ctx_id, metric_id, ident, name, descr FROM contextual_metrics WHERE id = $1;`
)

// Get returns a contextual metric if exists. Returns an error if fail to get.
func (c *ContextualMetrics) Get(ctx context.Context, id int64) (e bool, metric models.ContextualMetric, err error) {
	rows, err := c.Query(ctx, sqlCtxMetricsGet, id)
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
		e = true
		metric.Id = id
	}
	return e, metric, err
}

// MGet returns all metrics of a context with limit and offset. Returns an error if fails to get.
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

// ExistsIdent returns if a contextual metric's ident exists. Returns an error if fails to get.
func (c *ContextualMetrics) ExistsIdent(ctx context.Context, ident string, ctxId int32, id int64) (ie bool, err error) {
	err = c.QueryRow(ctx, sqlCtxMetricsExistsIdent, ident, ctxId, id).Scan(&ie)
	return ie, err
}

// GetMetricRequestByIdent returns the metric's information to make a data request by ident if exists.
// Returns an error if fails to get.
func (c *ContextualMetrics) GetMetricRequestByIdent(ctx context.Context, metricIdent string, contextIdent string, teamIdent string) (e bool, r models.MetricRequest, err error) {
	rows, err := c.Query(ctx, sqlGetIdContainerIdTeamId, teamIdent, contextIdent, metricIdent)
	if err != nil {
		return false, r, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(
			&r.MetricId,
			&r.MetricType,
			&r.ContainerId,
			&r.ContainerType,
		)
		if err != nil {
			return false, r, err
		}
		e = true
	}
	return e, r, nil
}

// ExistsContextAndIdent check if a context, metric and a contexual metric's ident exists.
// Returns an error if fails to check.
func (c *ContextualMetrics) ExistsContextMetricAndIdent(ctx context.Context, contextId int32, metricId int64, ident string, contextualMetricId int64) (ce bool, me bool, ie bool, err error) {
	rows, err := c.Query(ctx, sqlCtxMetricsExistsContextMetricAndIdent, contextId, metricId, ident, contextualMetricId)
	if err != nil {
		return false, false, false, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&ce, &me, &ie)
		if err != nil {
			return false, false, false, err
		}
	}
	return ce, me, ie, nil
}

// Create creates a new contextual metric. Returns an error if fails to create.
func (c *ContextualMetrics) Create(ctx context.Context, m models.ContextualMetric) error {
	_, err := c.Exec(ctx, sqlCtxMetricsCreate,
		m.ContextId,
		m.MetricId,
		m.Ident,
		m.Name,
		m.Descr,
	)
	return err
}

// Update updates a contextual metric if exists. Returns an error if fails to update.
func (c *ContextualMetrics) Update(ctx context.Context, m models.ContextualMetric) (e bool, err error) {
	t, err := c.Exec(ctx, sqlCtxMetricsUpdate,
		m.Ident,
		m.Name,
		m.Descr,
		m.Id,
	)
	return t.RowsAffected() != 0, err
}

// Delete deletes a contextual metric by id. Returns an error if fails to create.
func (c *ContextualMetrics) Delete(ctx context.Context, id int64) (e bool, err error) {
	t, err := c.Exec(ctx, sqlCtxMetricsDelete, id)
	return t.RowsAffected() != 0, err
}
