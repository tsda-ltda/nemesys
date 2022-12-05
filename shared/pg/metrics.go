package pg

import (
	"context"
	"database/sql"

	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/types"
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

type MetricsGetBasicResponse struct {
	// Exists is the metric existence.
	Exists bool
	// Metric is the metric.
	Metric models.Metric[struct{}]
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

type GetMetricRequestAndIntervalResult struct {
	// MetricRequest is the metric request.
	MetricRequest models.MetricRequest
	// Interval is the interval in seconds
	Interval int32
}

type GetMetricRequestResult struct {
	// Exists is the metric request existence.
	Exists bool
	// MetricRequest is the metric request.
	MetricRequest models.MetricRequest
	// Enabled is the metric enabled status.
	Enabled bool
}

type MetricDHSEnabledResult struct {
	// Exists is the metric request existence.
	Exists bool
	// Enabled is the metric dhs enabled.
	Enabled bool
}

const (
	sqlMetricsCreate = `INSERT INTO metrics 
		(container_id, container_type, name, descr, enabled, data_policy_id, rts_pulling_times, rts_data_cache_duration, dhs_enabled, dhs_interval, type, ev_expression)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING id;`
	sqlMetricsUpdate = `UPDATE metrics SET 
		(name, descr, enabled, data_policy_id, rts_pulling_times, rts_data_cache_duration, dhs_enabled, dhs_interval, type, ev_expression) 
		= ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $10) WHERE id = $11;`
	sqlMetricsGetRTSConfig = `SELECT rts_pulling_times, rts_data_cache_duration
		FROM metrics WHERE id = $1;`
	sqlMetricsExistsContainerAndDataPolicy = `SELECT 
		EXISTS (SELECT 1 FROM metrics WHERE id = $4),
		EXISTS (SELECT 1 FROM containers WHERE id = $1 AND type = $2),
		EXISTS (SELECT 1 FROM data_policies WHERE id = $3);`
	sqlMetricsGet = `SELECT 
		container_id, container_type, name, descr, enabled, data_policy_id, 
		rts_pulling_times, rts_data_cache_duration, dhs_enabled, dhs_interval, type, ev_expression FROM metrics WHERE id = $1;`
	sqlMetricsDelete                  = `DELETE FROM metrics WHERE id = $1;`
	sqlMetricsMGetSimplified          = `SELECT id, container_id, name, descr, enabled FROM metrics WHERE container_id = $1 AND container_type = $2 LIMIT $3 OFFSET $4;`
	sqlMetricsGetEvaluableExpression  = `SELECT ev_expression FROM metrics WHERE id = $1;`
	sqlMetricsGetEvaluableExpressions = `SELECT id, ev_expression FROM metrics WHERE id = ANY($1);`
	sqlMetricsEnabled                 = `WITH 
		m AS (SELECT enabled, container_id FROM metrics WHERE id = $1),
		c AS (SELECT enabled FROM containers WHERE id = (SELECT container_id FROM m))
		SELECT (SELECT enabled FROM m), (SELECT enabled FROM c);`
	sqlMetricsGetMetricsRequestsAndIntervals = `SELECT id, type, container_id, container_type, data_policy_id, dhs_interval FROM metrics WHERE dhs_enabled = true AND container_type != $1 LIMIT $2 OFFSET $3;`
	sqlMetricsGetRequest                     = `SELECT type, container_id, container_type, data_policy_id, enabled FROM metrics WHERE id = $1;`
	sqlMetricsDHSEnabled                     = `SELECT dhs_enabled FROM metrics WHERE id = $1;`
	sqlMetricsCountNonFlex                   = `SELECT COUNT(*) FROM metrics WHERE dhs_enabled = true AND container_type != $1;`
)

func (pg *PG) GetBasicMetric(ctx context.Context, id int64) (r MetricsGetBasicResponse, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlMetricsGet, id)
	if err != nil {
		return r, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(
			&r.Metric.Base.ContainerId,
			&r.Metric.Base.ContainerType,
			&r.Metric.Base.Name,
			&r.Metric.Base.Descr,
			&r.Metric.Base.Enabled,
			&r.Metric.Base.DataPolicyId,
			&r.Metric.Base.RTSPullingTimes,
			&r.Metric.Base.RTSCacheDuration,
			&r.Metric.Base.DHSEnabled,
			&r.Metric.Base.DHSInterval,
			&r.Metric.Base.Type,
			&r.Metric.Base.EvaluableExpression,
		)
		if err != nil {
			return r, err
		}
		r.Metric.Base.Id = id
		r.Exists = true
	}
	return r, nil
}

func (pg *PG) GetMetric(ctx context.Context, id int64) (r MetricsGetResponse, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlMetricsGet, id)
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
	metrics = []models.BaseMetricSimplified{}
	rows, err := pg.db.QueryContext(ctx, sqlMetricsMGetSimplified, containerId, containerType, limit, offset)
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

func (pg *PG) GetMetricRequest(ctx context.Context, id int64) (r GetMetricRequestResult, err error) {
	err = pg.db.QueryRowContext(ctx, sqlMetricsGetRequest, id).Scan(
		&r.MetricRequest.MetricType,
		&r.MetricRequest.ContainerId,
		&r.MetricRequest.ContainerType,
		&r.MetricRequest.DataPolicyId,
		&r.Enabled,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return r, nil
		}
		return r, err
	}
	r.MetricRequest.MetricId = id
	r.Exists = true
	return r, nil
}

func (pg *PG) GetMetricDHSEnabled(ctx context.Context, id int64) (r MetricDHSEnabledResult, err error) {
	err = pg.db.QueryRowContext(ctx, sqlMetricsDHSEnabled, id).Scan(&r.Enabled)
	if err != nil {
		if err == sql.ErrNoRows {
			return r, nil
		}
		return r, err
	}
	r.Exists = true
	return r, nil
}

func (pg *PG) createMetric(ctx context.Context, tx *sql.Tx, metric models.BaseMetric) (id int64, err error) {
	err = tx.QueryRowContext(ctx, sqlMetricsCreate,
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

func (pg *PG) CreateBasicMetric(ctx context.Context, metric models.Metric[struct{}]) (id int64, err error) {
	err = pg.db.QueryRowContext(ctx, sqlMetricsCreate,
		metric.Base.ContainerId,
		metric.Base.ContainerType,
		metric.Base.Name,
		metric.Base.Descr,
		metric.Base.Enabled,
		metric.Base.DataPolicyId,
		metric.Base.RTSPullingTimes,
		metric.Base.RTSCacheDuration,
		metric.Base.DHSEnabled,
		metric.Base.DHSInterval,
		metric.Base.Type,
		metric.Base.EvaluableExpression,
	).Scan(&id)
	return id, err
}

func (pg *PG) UpdateBasicMetric(ctx context.Context, metric models.Metric[struct{}]) (exists bool, err error) {
	t, err := pg.db.ExecContext(ctx, sqlMetricsUpdate,
		metric.Base.Name,
		metric.Base.Descr,
		metric.Base.Enabled,
		metric.Base.DataPolicyId,
		metric.Base.RTSPullingTimes,
		metric.Base.RTSCacheDuration,
		metric.Base.DHSEnabled,
		metric.Base.DHSInterval,
		metric.Base.Type,
		metric.Base.EvaluableExpression,
		metric.Base.Id,
	)
	if err != nil {
		return false, err
	}
	rowsAffected, _ := t.RowsAffected()
	return rowsAffected != 0, err
}

func (pg *PG) updateMetric(ctx context.Context, tx *sql.Tx, metric models.BaseMetric) (exists bool, err error) {
	t, err := tx.ExecContext(ctx, sqlMetricsUpdate,
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
	if err != nil {
		return false, err
	}
	rowsAffected, _ := t.RowsAffected()
	return rowsAffected != 0, err
}

func (pg *PG) DeleteMetric(ctx context.Context, id int64) (exists bool, err error) {
	t, err := pg.db.ExecContext(ctx, sqlMetricsDelete, id)
	if err != nil {
		return false, err
	}
	rowsAffected, _ := t.RowsAffected()
	return rowsAffected != 0, err
}

func (pg *PG) GetMetricEvaluableExpression(ctx context.Context, id int64) (r MetricsGetEvaluableExpressionResponse, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlMetricsGetEvaluableExpression, id)
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

func (pg *PG) GetMetricsEvaluableExpressions(ctx context.Context, ids []int64) (expressions []models.MetricEvaluableExpression, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlMetricsGetEvaluableExpressions, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	expressions = make([]models.MetricEvaluableExpression, 0, len(ids))
	for rows.Next() {
		var e models.MetricEvaluableExpression
		var err = rows.Scan(&e.Id, &e.Expression)
		if err != nil {
			return nil, err
		}
		expressions = append(expressions, e)
	}
	return expressions, err
}

func (pg *PG) GetMetricRTSConfig(ctx context.Context, id int64) (r MetricsGetRTSConfigResponse, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlMetricsGetRTSConfig, id)
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
	return r, pg.db.QueryRowContext(ctx, sqlMetricsExistsContainerAndDataPolicy, base.ContainerId, base.ContainerType, base.DataPolicyId, base.Id).Scan(
		&r.Exists,
		&r.ContainerExists,
		&r.DataPolicyExists,
	)
}

func (pg *PG) MetricEnabled(ctx context.Context, id int32) (r MetricsEnabledResponse, err error) {
	var ce *bool
	var me *bool
	rows, err := pg.db.QueryContext(ctx, sqlContainersEnabled, id)
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

func (pg *PG) GetMetricsRequestsAndIntervals(ctx context.Context, limit int, offset int) (r []GetMetricRequestAndIntervalResult, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlMetricsGetMetricsRequestsAndIntervals, types.CTFlexLegacy, limit, offset)
	if err != nil {
		return r, err
	}
	defer rows.Close()
	r = make([]GetMetricRequestAndIntervalResult, 0, limit)
	for rows.Next() {
		var result GetMetricRequestAndIntervalResult
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
		r = append(r, result)
	}
	return r, nil
}

func (pg *PG) CountNonFlexMetrics(ctx context.Context) (n int, err error) {
	return n, pg.db.QueryRowContext(ctx, sqlMetricsCountNonFlex, types.CTFlexLegacy).Scan(&n)
}
