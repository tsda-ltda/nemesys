package pg

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/types"
	"github.com/jackc/pgx/v5"
)

type MetricsGetEvaluableExpressionResponse struct {
	// Exists is the metric existence.
	Exists bool
	// Expression is the evaluable expression.
	Expression string
}

type MetricsGetResponse struct {
	// Exists is the metric existence.
	Exists bool
	// Metric is the metric.
	Metric models.BaseMetric
}

type MetricsGetRTSConfigResponse struct {
	// Exists is the metric existence.
	Exists bool
	// RTSConfig is the metric RTS configuration.
	RTSConfig models.RTSMetricConfig
}

type MetricsExistsContainerAndDataPolicyResponse struct {
	// Exists is the metric existence.
	Exists bool
	// ContainerExists is the container existence.
	ContainerExists bool
	// DataPolicyExists is the data policy existence.
	DataPolicyExists bool
}

type MetricsEnabledResponse struct {
	// Exists is the metric existence.
	Exists bool
	// ContainerExists is the container existence.
	ContainerExists bool
	// Enabled is the metric's enabled status.
	Enabled bool
	// ContainerEnabled is the container's enabled status.
	ContainerEnabled bool
}

type GetMetricsRequestsAndIntervalsResult struct {
	// MetricRequest is the metric request.
	MetricRequest models.MetricRequest
	// Interval is the interval in seconds
	Interval int32
}

const (
	sqlMetricsCreate = `INSERT INTO metrics 
		(container_id, container_type, name, descr, enabled, data_policy_id, rts_pulling_times, rts_data_cache_duration, dhs_enabled, dhs_interval, type, ev_expression)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING id;`
	sqlMetricsUpdate = `UPDATE metrics SET 
		(name, descr, enabled, data_policy_id, rts_pulling_times, rts_data_cache_duration, dhs_enabled, dhs_interval, type, ev_expression) 
		= ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) WHERE id = $11;`
	sqlMetricsGetRTSConfig = `SELECT rts_pulling_times, rts_data_cache_duration
		FROM metrics WHERE id = $1;`
	sqlMetricsExistsContainerAndDataPolicy = `SELECT 
		EXISTS (SELECT 1 FROM metrics WHERE id = $4),
		EXISTS (SELECT 1 FROM containers WHERE id = $1 AND type = $2),
		EXISTS (SELECT 1 FROM data_policies WHERE id = $3);`
	sqlMetricsGet = `SELECT 
		container_id, container_type, name, descr, enabled, data_policy_id, 
		rts_pulling_times, rts_data_cache_duration, dhs_enabled, dhs_interval, type, ev_expression FROM metrics WHERE id = $1;`
	sqlMetricsDelete                 = `DELETE FROM metrics WHERE id = $1;`
	sqlMetricsMGetSimplified         = `SELECT id, container_id, name, descr, enabled FROM metrics WHERE container_id = $1 AND container_type = $2 LIMIT $3 OFFSET $4;`
	sqlMetricsGetEvaluableExpression = `SELECT ev_expression FROM metrics WHERE id = $1;`
	sqlMetricsEnabled                = `WITH 
		m AS (SELECT enabled, container_id FROM metrics WHERE id = $1),
		c AS (SELECT enabled FROM containers WHERE id = (SELECT container_id FROM m))
		SELECT (SELECT enabled FROM m), (SELECT enabled FROM c);`
	sqlMetricsGetMetricsRequestsAndIntervals = `SELECT id, type, container_id, container_type, data_policy_id, dhs_interval FROM metrics WHERE dhs_enabled = true LIMIT $1 OFFSET $2;`
)

func (pg *PG) GetMetric(ctx context.Context, id int64) (r MetricsGetResponse, err error) {
	c, err := pg.pool.Acquire(ctx)
	if err != nil {
		return r, err
	}
	defer c.Release()
	rows, err := c.Query(ctx, sqlMetricsGet, id)
	if err != nil {
		return r, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(
			&r.Metric.ContainerId,
			&r.Metric.ContainerType,
			&r.Metric.Name,
			&r.Metric.Descr,
			&r.Metric.Enabled,
			&r.Metric.DataPolicyId,
			&r.Metric.RTSPullingTimes,
			&r.Metric.RTSCacheDuration,
			&r.Metric.DHSEnabled,
			&r.Metric.DHSInterval,
			&r.Metric.Type,
			&r.Metric.EvaluableExpression,
		)
		if err != nil {
			return r, err
		}
		r.Metric.Id = id
		r.Exists = true
	}
	return r, nil
}

func (pg *PG) GetMetricsSimplified(ctx context.Context, containerType types.ContainerType, containerId int32, limit int, offset int) (metrics []models.BaseMetricSimplified, err error) {
	c, err := pg.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Release()
	metrics = []models.BaseMetricSimplified{}
	rows, err := c.Query(ctx, sqlMetricsMGetSimplified, containerId, containerType, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var m models.BaseMetricSimplified
		err = rows.Scan(
			&m.Id,
			&m.ContainerId,
			&m.Name,
			&m.Descr,
			&m.Enabled,
		)
		if err != nil {
			return nil, err
		}
		m.ContainerType = containerType
		metrics = append(metrics, m)
	}
	return metrics, nil
}

func (pg *PG) createMetric(ctx context.Context, tx pgx.Tx, metric models.BaseMetric) (id int64, err error) {
	err = tx.QueryRow(ctx, sqlMetricsCreate,
		metric.ContainerId,
		metric.ContainerType,
		metric.Name,
		metric.Descr,
		metric.Enabled,
		metric.DataPolicyId,
		metric.RTSPullingTimes,
		metric.RTSCacheDuration,
		metric.DHSEnabled,
		metric.DHSInterval,
		metric.Type,
		metric.EvaluableExpression,
	).Scan(&id)
	return id, err
}

func (pg *PG) updateMetric(ctx context.Context, tx pgx.Tx, metric models.BaseMetric) (exists bool, err error) {
	t, err := tx.Exec(ctx, sqlMetricsUpdate,
		metric.Name,
		metric.Descr,
		metric.Enabled,
		metric.DataPolicyId,
		metric.RTSPullingTimes,
		metric.RTSCacheDuration,
		metric.DHSEnabled,
		metric.DHSInterval,
		metric.Type,
		metric.EvaluableExpression,
		metric.Id,
	)
	return t.RowsAffected() != 0, err
}

func (pg *PG) DeleteMetric(ctx context.Context, id int64) (exists bool, err error) {
	c, err := pg.pool.Acquire(ctx)
	if err != nil {
		return false, err
	}
	defer c.Release()
	t, err := c.Exec(ctx, sqlMetricsDelete, id)
	return t.RowsAffected() != 0, err
}

func (pg *PG) GetMetricEvaluableExpression(ctx context.Context, id int64) (r MetricsGetEvaluableExpressionResponse, err error) {
	c, err := pg.pool.Acquire(ctx)
	if err != nil {
		return r, err
	}
	defer c.Release()
	rows, err := c.Query(ctx, sqlMetricsGetEvaluableExpression, id)
	if err != nil {
		return r, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&r.Expression)
		if err != nil {
			return r, err
		}
		r.Exists = true
	}
	return r, err
}

func (pg *PG) GetMetricRTSConfig(ctx context.Context, id int64) (r MetricsGetRTSConfigResponse, err error) {
	c, err := pg.pool.Acquire(ctx)
	if err != nil {
		return r, err
	}
	defer c.Release()
	rows, err := c.Query(ctx, sqlMetricsGetRTSConfig, id)
	if err != nil {
		return r, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(
			&r.RTSConfig.PullingTimes,
			&r.RTSConfig.CacheDuration,
		)
		if err != nil {
			return r, err
		}
		r.Exists = true
	}
	return r, nil
}

func (pg *PG) MetricContainerAndDataPolicyExists(ctx context.Context, base models.BaseMetric) (r MetricsExistsContainerAndDataPolicyResponse, err error) {
	c, err := pg.pool.Acquire(ctx)
	if err != nil {
		return r, err
	}
	defer c.Release()
	return r, c.QueryRow(ctx, sqlMetricsExistsContainerAndDataPolicy, base.ContainerId, base.ContainerType, base.DataPolicyId, base.Id).Scan(
		&r.Exists,
		&r.ContainerExists,
		&r.DataPolicyExists,
	)
}

func (pg *PG) MetricEnabled(ctx context.Context, id int32) (r MetricsEnabledResponse, err error) {
	c, err := pg.pool.Acquire(ctx)
	if err != nil {
		return r, err
	}
	defer c.Release()
	var ce *bool
	var me *bool
	rows, err := c.Query(ctx, sqlContainersEnabled, id)
	if err != nil {
		return r, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&me, &ce)
		if err != nil {
			return r, err
		}
	}
	if ce != nil {
		r.ContainerExists = true
		r.ContainerEnabled = *ce
	}
	if me != nil {
		r.Exists = true
		r.Enabled = *me
	}

	return r, nil
}

func (pg *PG) GetMetricsRequestsAndIntervals(ctx context.Context, limit int, offset int) (r []GetMetricsRequestsAndIntervalsResult, err error) {
	c, err := pg.pool.Acquire(ctx)
	if err != nil {
		return r, err
	}
	defer c.Release()
	rows, err := c.Query(ctx, sqlMetricsGetMetricsRequestsAndIntervals, limit, offset)
	if err != nil {
		return r, err
	}
	defer rows.Close()
	results := make([]GetMetricsRequestsAndIntervalsResult, 0, limit)
	for rows.Next() {
		var result GetMetricsRequestsAndIntervalsResult
		err = rows.Scan(
			&result.MetricRequest.MetricId,
			&result.MetricRequest.MetricType,
			&result.MetricRequest.ContainerId,
			&result.MetricRequest.ContainerType,
			&result.MetricRequest.DataPolicyId,
			&result.Interval,
		)
		if err != nil {
			return r, err
		}
		results = append(results, result)
	}
	r = results
	return r, nil
}
