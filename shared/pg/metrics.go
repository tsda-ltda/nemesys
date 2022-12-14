package pg

import (
	"context"
	"database/sql"

	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/types"
)

var BasicMetricValidOrderByColumns = []string{"name", "descr"}

type BasicMetricQueryFilters struct {
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

func (f BasicMetricQueryFilters) GetOrderBy() string {
	return f.OrderBy
}

func (f BasicMetricQueryFilters) GetOrderByFn() string {
	return f.OrderByFn
}

func (f BasicMetricQueryFilters) GetLimit() int {
	return f.Limit
}

func (f BasicMetricQueryFilters) GetOffset() int {
	return f.Offset
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
	sqlMetricsDelete                  = `DELETE FROM metrics WHERE id = $1;`
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
	sqlMetricsGetAlarmExpressions            = `SELECT e.id, e.expression, e.category_id FROM alarm_expressions e
	LEFT JOIN metrics_alarm_expressions_rel r ON r.expression_id = e.id WHERE r.metric_id = $1;`
	sqlMetricsGetAlarmsExpressions = `SELECT r.metric_id, e.id, e.expression, e.category_id FROM alarm_expressions e
	FULL OUTER JOIN metrics_alarm_expressions_rel r ON r.expression_id = e.id WHERE r.metric_id = ANY ($1);`

	customSqlBasicMetricsMGet = `SELECT id, name, descr, enabled, data_policy_id, 
	rts_pulling_times, rts_data_cache_duration, dhs_enabled, dhs_interval, type, ev_expression`
)

func (pg *PG) GetBasicMetric(ctx context.Context, id int64) (exists bool, metric models.Metric[struct{}], err error) {
	rows, err := pg.db.QueryContext(ctx, sqlMetricsGet, id)
	if err != nil {
		return false, metric, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(
			&metric.Base.ContainerId,
			&metric.Base.ContainerType,
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
		)
		if err != nil {
			return false, metric, err
		}
		metric.Base.Id = id
		exists = true
	}
	return exists, metric, nil
}

func (pg *PG) GetMetric(ctx context.Context, id int64) (exists bool, metric models.BaseMetric, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlMetricsGet, id)
	if err != nil {
		return false, metric, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(
			&metric.ContainerId,
			&metric.ContainerType,
			&metric.Name,
			&metric.Descr,
			&metric.Enabled,
			&metric.DataPolicyId,
			&metric.RTSPullingTimes,
			&metric.RTSCacheDuration,
			&metric.DHSEnabled,
			&metric.DHSInterval,
			&metric.Type,
			&metric.EvaluableExpression,
		)
		if err != nil {
			return false, metric, err
		}
		metric.Id = id
		exists = true
	}
	return exists, metric, nil
}

func (pg *PG) GetBasicMetrics(ctx context.Context, filters BasicMetricQueryFilters) (metrics []models.Metric[struct{}], err error) {
	filters.ContainerType = types.CTBasic
	sql, params, err := applyFilters(filters, customSqlBasicMetricsMGet, BasicMetricValidOrderByColumns)
	if err != nil {
		return nil, err
	}
	rows, err := pg.db.QueryContext(ctx, sql, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	metrics = make([]models.Metric[struct{}], 0, filters.Limit)
	var m models.Metric[struct{}]
	m.Base.ContainerId = filters.ContainerId
	m.Base.ContainerType = filters.ContainerType
	for rows.Next() {
		err = rows.Scan(
			&m.Base.Id,
			&m.Base.Name,
			&m.Base.Descr,
			&m.Base.Enabled,
			&m.Base.DataPolicyId,
			&m.Base.RTSPullingTimes,
			&m.Base.RTSCacheDuration,
			&m.Base.DHSEnabled,
			&m.Base.DHSInterval,
			&m.Base.Type,
			&m.Base.EvaluableExpression,
		)
		if err != nil {
			return nil, err
		}
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

func (pg *PG) GetMetricDHSEnabled(ctx context.Context, id int64) (exists bool, enabled bool, err error) {
	err = pg.db.QueryRowContext(ctx, sqlMetricsDHSEnabled, id).Scan(&enabled)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, false, nil
		}
		return false, false, err
	}
	return true, enabled, nil
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

func (pg *PG) GetMetricEvaluableExpression(ctx context.Context, id int64) (exists bool, expression string, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlMetricsGetEvaluableExpression, id)
	if err != nil {
		return false, expression, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&expression)
		if err != nil {
			return false, expression, err
		}
		exists = true
	}
	return exists, expression, err
}

func (pg *PG) GetMetricsEvaluableExpressions(ctx context.Context, ids []int64) (expressions []models.MetricEvaluableExpression, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlMetricsGetEvaluableExpressions, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	expressions = make([]models.MetricEvaluableExpression, 0, len(ids))
	var e models.MetricEvaluableExpression
	for rows.Next() {
		var err = rows.Scan(&e.Id, &e.Expression)
		if err != nil {
			return nil, err
		}
		expressions = append(expressions, e)
	}
	return expressions, err
}

func (pg *PG) GetMetricRTSConfig(ctx context.Context, id int64) (exists bool, RTSConfig models.RTSMetricConfig, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlMetricsGetRTSConfig, id)
	if err != nil {
		return exists, RTSConfig, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(
			&RTSConfig.PullingTimes,
			&RTSConfig.CacheDuration,
		)
		if err != nil {
			return false, RTSConfig, err
		}
		exists = true
	}
	return exists, RTSConfig, nil
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
	var result GetMetricRequestAndIntervalResult
	for rows.Next() {
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

func (pg *PG) GetMetricAlarmExpressions(ctx context.Context, id int64) (expressions []models.AlarmExpressionSimplified, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlMetricsGetAlarmExpressions, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	expressions = []models.AlarmExpressionSimplified{}
	var exp models.AlarmExpressionSimplified
	for rows.Next() {
		err = rows.Scan(
			&exp.Id,
			&exp.Expression,
			&exp.AlarmCategoryId,
		)
		if err != nil {
			return nil, err
		}
		expressions = append(expressions, exp)
	}
	return expressions, err
}

func (pg *PG) GetMetricsAlarmExpressions(ctx context.Context, ids []int64) (expressions [][]models.AlarmExpressionSimplified, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlMetricsGetAlarmsExpressions, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	expressions = make([][]models.AlarmExpressionSimplified, len(ids))
	var exp models.AlarmExpressionSimplified
	var metricId int64
	for rows.Next() {
		err = rows.Scan(
			&metricId,
			&exp.Id,
			&exp.Expression,
			&exp.AlarmCategoryId,
		)
		if err != nil {
			return nil, err
		}
		for i, id := range ids {
			if id == metricId {
				expressions[i] = append(expressions[i], exp)
			}
		}
	}
	return expressions, err
}
