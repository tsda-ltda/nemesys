package db

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/types"
	"github.com/jackc/pgx/v5"
)

type Metrics struct {
	*pgx.Conn
}

const (
	sqlMetricsCreate = `INSERT INTO metrics 
		(container_id, container_type, name, descr, data_policy_id, rts_pulling_times, rts_cache_duration, type, ev_expression)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id;`
	sqlMetricsUpdate = `UPDATE metrics SET 
		(name, descr, data_policy_id, rts_pulling_times, rts_cache_duration, type, ev_expression) 
		= ($1, $2, $3, $4, $5, $6, $7) WHERE id = $8;`
	sqlMetricsGetRTSConfig = `SELECT rts_pulling_times, rts_cache_duration
		FROM metrics WHERE id = $1;`
	sqlMetricsExistsContainerAndDataPolicy = `SELECT 
		EXISTS (SELECT 1 FROM metrics WHERE id = $4),
		EXISTS (SELECT 1 FROM containers WHERE id = $1 AND type = $2),
		EXISTS (SELECT 1 FROM data_policies WHERE id = $3);`
	sqlMetricsGet = `SELECT 
		container_id, container_type, name, descr, data_policy_id, 
		rts_pulling_times, rts_cache_duration, type, ev_expression FROM metrics WHERE id = $1;`
	sqlMetricsDelete                 = `DELETE FROM metrics WHERE id = $1;`
	sqlMetricsMGetSimplified         = `SELECT id, container_id, name, descr FROM metrics WHERE container_id = $1 LIMIT $2 OFFSET $3;`
	sqlMetricsGetEvaluableExpression = `SELECT ev_expression FROM metrics WHERE id = $1;`
)

// GetEvaluableExpression returns the metric evaluable expression if exists. Returns an error if fails to get.
func (c *Metrics) GetEvaluableExpression(ctx context.Context, id int64) (e bool, expression string, err error) {
	rows, err := c.Query(ctx, sqlMetricsGetEvaluableExpression, id)
	if err != nil {
		return false, expression, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&expression)
		if err != nil {
			return false, expression, err
		}
		e = true
	}
	return e, expression, err
}

// Get returns a metric if exists. Returns an error if fails to get.
func (c *Metrics) Get(ctx context.Context, id int64) (e bool, m models.BaseMetric, err error) {
	rows, err := c.Query(ctx, sqlMetricsGet, id)
	if err != nil {
		return false, m, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(
			&m.ContainerId,
			&m.ContainerType,
			&m.Name,
			&m.Descr,
			&m.DataPolicyId,
			&m.RTSPullingTimes,
			&m.RTSCacheDuration,
			&m.Type,
			&m.EvaluableExpression,
		)
		if err != nil {
			return false, m, err
		}
		m.Id = id
		e = true
	}
	return e, m, nil
}

// MGetSimplified returns all simplified metrics of a container type with a limit and offset. Returns an error if fail to get.
func (c *Metrics) MGetSimplified(ctx context.Context, containerType types.ContainerType, containerId int32, limit int, offset int) (metrics []models.BaseMetricSimplified, err error) {
	metrics = []models.BaseMetricSimplified{}
	rows, err := c.Query(ctx, sqlMetricsMGetSimplified, containerId, limit, offset)
	if err != nil {
		return metrics, err
	}
	defer rows.Close()
	for rows.Next() {
		var m models.BaseMetricSimplified
		err = rows.Scan(
			&m.Id,
			&m.ContainerId,
			&m.Name,
			&m.Descr,
		)
		if err != nil {
			return metrics, err
		}
		m.ContainerType = containerType
		metrics = append(metrics, m)
	}
	return metrics, nil
}

// Create creates a metric. Returns an error if fails to create.
func (c *Metrics) Create(ctx context.Context, metric models.BaseMetric) (id int64, err error) {
	err = c.QueryRow(ctx, sqlMetricsCreate,
		metric.ContainerId,
		metric.ContainerType,
		metric.Name,
		metric.Descr,
		metric.DataPolicyId,
		metric.RTSPullingTimes,
		metric.RTSCacheDuration,
		metric.Type,
		metric.EvaluableExpression,
	).Scan(&id)
	return id, err
}

// Update updates a metric if exists. Returns an error if fails to update.
func (c *Metrics) Update(ctx context.Context, metric models.BaseMetric) (e bool, err error) {
	t, err := c.Exec(ctx, sqlMetricsUpdate,
		metric.Name,
		metric.Descr,
		metric.DataPolicyId,
		metric.RTSPullingTimes,
		metric.RTSCacheDuration,
		metric.Type,
		metric.EvaluableExpression,
		metric.Id,
	)
	return t.RowsAffected() != 0, err
}

// Delete deletes a metric if exists. Returns an error if fails to delete.
func (c *Metrics) Delete(ctx context.Context, id int64) (e bool, err error) {
	t, err := c.Exec(ctx, sqlMetricsDelete, id)
	return t.RowsAffected() != 0, err
}

// GetRTSConfig returns the metric's RTS configuration. Returns an error if fail to get.
func (c *Metrics) GetRTSConfig(ctx context.Context, id int64) (e bool, info models.RTSMetricConfig, err error) {
	rows, err := c.Query(ctx, sqlMetricsGetRTSConfig, id)
	if err != nil {
		return false, info, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(
			&info.PullingTimes,
			&info.CacheDuration,
		)
		if err != nil {
			return false, info, err
		}
		e = true
	}
	return e, info, nil
}

// ExistsIdentAndContainerAndDataPolicy check if exists a diferent metric with the ident and if a container with the type exists.
// Returns an error if fail to check.
func (c *Metrics) ExistsContainerAndDataPolicy(ctx context.Context, containerId int32, containerType types.ContainerType, dataPolicyId int16, id int64) (e bool, ce bool, dpe bool, err error) {
	err = c.QueryRow(ctx, sqlMetricsExistsContainerAndDataPolicy, containerId, containerType, dataPolicyId, id).Scan(&e, &ce, &dpe)
	return e, ce, dpe, err
}
