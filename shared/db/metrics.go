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
		(container_id, container_type, name, ident, descr, data_policy_id, rts_pulling_interval, rts_pulling_times, rts_cache_duration)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id;`
	sqlMetricsUpdate = `UPDATE metrics SET 
		(name, ident, descr, data_policy_id, rts_pulling_interval, rts_pulling_times, rts_cache_duration) 
		= ($1, $2, $3, $4, $5, $6, $7) WHERE id = $8;`
	sqlMetricsGetRTSConfig = `SELECT rts_pulling_interval, rts_pulling_times, rts_cache_duration 
		FROM metrics WHERE id = $1;`
	sqlMetricsExistsIdentAndContainerAndDataPolicy = `SELECT 
		EXISTS (SELECT 1 FROM metrics WHERE id = $5),
		EXISTS (SELECT 1 FROM containers WHERE id = $1 AND type = $2),
		EXISTS (SELECT 1 FROM data_policies WHERE id = $3),
		EXISTS (SELECT 1 FROM metrics WHERE ident = $4 AND container_id = $1 AND id != $5);`
	sqlMetricsGet = `SELECT 
		container_id, container_type, name, ident, descr, data_policy_id, 
		rts_pulling_interval, rts_pulling_times, rts_cache_duration FROM metrics WHERE id = $1;`
	sqlMetricsDelete         = `DELETE FROM metrics WHERE id = $1;`
	sqlMetricsMGetSimplified = `SELECT id, container_id, name, ident, descr FROM metrics WHERE container_type = $1 LIMIT $2 OFFSET $3;`
)

func (c *Metrics) Get(ctx context.Context, id int) (e bool, m models.BaseMetric, err error) {
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
			&m.Ident,
			&m.Descr,
			&m.DataPolicyId,
			&m.RTSPullingInterval,
			&m.RTSPullingTimes,
			&m.RTSCacheDuration,
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
func (c *Metrics) MGetSimplified(ctx context.Context, t types.ContainerType, limit int, offset int) (metrics []models.BaseMetricSimplified, err error) {
	metrics = []models.BaseMetricSimplified{}
	rows, err := c.Query(ctx, sqlMetricsMGetSimplified, t, limit, offset)
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
			&m.Ident,
			&m.Descr,
		)
		if err != nil {
			return metrics, err
		}
		m.ContainerType = t
		metrics = append(metrics, m)
	}
	return metrics, nil
}

// Create creates a metric. Returns an error if fails to create.
func (c *Metrics) Create(ctx context.Context, metric models.BaseMetric) (id int, err error) {
	err = c.QueryRow(ctx, sqlMetricsCreate,
		metric.ContainerId,
		metric.ContainerType,
		metric.Name,
		metric.Ident,
		metric.Descr,
		metric.DataPolicyId,
		metric.RTSPullingInterval,
		metric.RTSPullingTimes,
		metric.RTSCacheDuration,
	).Scan(&id)
	return id, err
}

// Update updates a metric if exists. Returns an error if fails to update.
func (c *Metrics) Update(ctx context.Context, metric models.BaseMetric) (e bool, err error) {
	t, err := c.Exec(ctx, sqlMetricsUpdate,
		metric.Name,
		metric.Ident,
		metric.Descr,
		metric.DataPolicyId,
		metric.RTSPullingInterval,
		metric.RTSPullingTimes,
		metric.RTSCacheDuration,
		metric.Id,
	)
	return t.RowsAffected() != 0, err
}

// Delete deletes a metric if exists. Returns an error if fails to delete.
func (c *Metrics) Delete(ctx context.Context, id int) (e bool, err error) {
	t, err := c.Exec(ctx, sqlMetricsDelete, id)
	return t.RowsAffected() != 0, err
}

// GetRTSConfig returns the metric RTS configuration. Returns an error if fail to get.
func (c *Metrics) GetRTSConfig(ctx context.Context, id int) (e bool, info models.RTSMetricInfo, err error) {
	rows, err := c.Query(ctx, sqlMetricsGetRTSConfig, id)
	if err != nil {
		return false, info, err
	}
	for rows.Next() {
		err = rows.Scan(&info.PullingInterval, &info.PullingTimes, &info.CacheDuration)
		if err != nil {
			return false, info, err
		}
		e = true
	}
	return e, info, nil
}

// ExistsIdentAndContainerAndDataPolicy check if exists a diferent metric with the ident and if a container with the type exists.
// Returns an error if fail to check.
func (c *Metrics) ExistsIdentAndContainerAndDataPolicy(ctx context.Context, containerId int, containerType types.ContainerType, dataPolicyId int, ident string, id int) (e bool, ce bool, dpe bool, ie bool, err error) {
	err = c.QueryRow(ctx, sqlMetricsExistsIdentAndContainerAndDataPolicy, containerId, containerType, dataPolicyId, ident, id).Scan(&e, &ce, &dpe, &ie)
	return e, ce, dpe, ie, err
}
