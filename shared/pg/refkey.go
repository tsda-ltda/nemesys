package pg

import (
	"context"
	"database/sql"

	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/types"
)

const (
	sqlRefkeyCreate           = `INSERT INTO metrics_ref (refkey, metric_id) VALUES ($1, $2) RETURNING id;`
	sqlRefkeyUpdate           = `UPDATE metrics_ref SET (refkey, metric_id) = ($1, $2) WHERE id = $3;`
	sqlRefKeyDelete           = `DELETE FROM metrics_ref WHERE id = $1;`
	sqlRefkeyGetByRefkey      = `SELECT id, metric_id FROM metrics_ref WHERE refkey = $1;`
	sqlRefkeyGetMetricRefKeys = `SELECT id, refkey FROM metrics_ref WHERE metric_id = $1;`
	sqlRefkeyExists           = `SELECT 
		EXISTS (SELECT 1 FROM metrics WHERE id = $1 AND container_type = $2),
		EXISTS (SELECT 1 FROM metrics_ref WHERE refkey = $3 and id != $4);`
)

func (pg *PG) CreateMetricRefkey(ctx context.Context, rk models.MetricRefkey) (id int64, err error) {
	return id, pg.db.QueryRowContext(ctx, sqlRefkeyCreate, rk.Refkey, rk.MetricId).Scan(&id)
}

func (pg *PG) UpdateMetricRefkey(ctx context.Context, rk models.MetricRefkey) (exists bool, err error) {
	t, err := pg.db.ExecContext(ctx, sqlRefkeyUpdate, rk.Refkey, rk.MetricId, rk.Id)
	if err != nil {
		return false, err
	}
	rowsAffected, _ := t.RowsAffected()
	return rowsAffected != 0, err
}

func (pg *PG) DeleteMetricRefkey(ctx context.Context, id int64) (exists bool, err error) {
	t, err := pg.db.ExecContext(ctx, sqlRefKeyDelete, id)
	rowsAffected, _ := t.RowsAffected()
	if err != nil {
		return false, err
	}
	return rowsAffected != 0, err
}

func (pg *PG) GetRefkey(ctx context.Context, refkey string) (exists bool, rk models.MetricRefkey, err error) {
	err = pg.db.QueryRowContext(ctx, sqlRefkeyGetByRefkey, refkey).Scan(
		&rk.Id,
		&rk.MetricId,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, rk, nil
		}
		return false, rk, err
	}
	rk.Refkey = refkey
	return true, rk, nil
}

func (pg *PG) GetMetricRefkeys(ctx context.Context, metricId int64) (rks []models.MetricRefkey, err error) {
	rks = []models.MetricRefkey{}
	rows, err := pg.db.QueryContext(ctx, sqlRefkeyGetMetricRefKeys, metricId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var rk models.MetricRefkey
		err = rows.Scan(&rk.Id, &rk.Refkey)
		if err != nil {
			return nil, err
		}
		rk.MetricId = metricId
		rks = append(rks, rk)
	}
	return rks, nil
}

func (pg *PG) MetricAndRefkeyExists(ctx context.Context, metricId int64, containerType types.ContainerType, refkey string, id int64) (metricExists bool, refkeyExists bool, err error) {
	return metricExists, refkeyExists, pg.db.QueryRowContext(ctx, sqlRefkeyExists, metricId, containerType, refkey, id).Scan(&metricExists, &refkeyExists)
}
