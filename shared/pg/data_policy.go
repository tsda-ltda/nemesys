package pg

import (
	"context"
	"database/sql"

	"github.com/fernandotsda/nemesys/shared/models"
)

const (
	sqlDPCreate = `INSERT INTO data_policies (name, descr, use_aggr, retention, aggr_retention, aggr_interval, aggr_fn) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id;`
	sqlDPUpdate = `UPDATE data_policies SET (name, descr, retention, use_aggr, aggr_retention, aggr_interval, aggr_fn) = ($1, $2, $3, $4, $5, $6, $7) WHERE id = $8;`
	sqlDPDelete = `DELETE FROM data_policies WHERE id = $1;`
	sqlDPGet    = `SELECT id, name, descr, retention, use_aggr, aggr_retention, aggr_interval, aggr_fn FROM data_policies WHERE id = $1;`
	sqlDPMGet   = `SELECT id, name, descr, retention, use_aggr, aggr_retention, aggr_interval, aggr_fn FROM data_policies;`
	sqlDPCount  = `SELECT COUNT(*) FROM data_policies;`
)

func (pg *PG) CountDataPolicy(ctx context.Context) (n int64, err error) {
	err = pg.db.QueryRowContext(ctx, sqlDPCount).Scan(&n)
	return n, err
}

func (pg *PG) CreateDataPolicy(ctx context.Context, dp models.DataPolicy) (tx *sql.Tx, id int16, err error) {
	c, err := pg.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, id, err
	}
	err = c.QueryRowContext(ctx, sqlDPCreate,
		dp.Name,
		dp.Descr,
		dp.UseAggr,
		dp.Retention,
		dp.AggrRetention,
		dp.AggrInterval,
		dp.AggrFn,
	).Scan(&id)
	if err != nil {
		return nil, id, err
	}
	return c, id, nil
}

func (pg *PG) UpdateDataPolicy(ctx context.Context, dp models.DataPolicy) (tx *sql.Tx, exists bool, err error) {
	c, err := pg.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, false, err
	}
	t, err := c.ExecContext(ctx, sqlDPUpdate,
		dp.Name,
		dp.Descr,
		dp.Retention,
		dp.UseAggr,
		dp.AggrRetention,
		dp.AggrInterval,
		dp.AggrFn,
		dp.Id,
	)
	if err != nil {
		return nil, false, err
	}
	rowsAffected, _ := t.RowsAffected()
	return c, rowsAffected != 0, err
}

func (pg *PG) DeleteDataPolicy(ctx context.Context, id int16) (exists bool, err error) {
	t, err := pg.db.ExecContext(ctx, sqlDPDelete, id)
	if err != nil {
		return false, err
	}
	rowsAffected, _ := t.RowsAffected()
	return rowsAffected != 0, err
}

func (pg *PG) GetDataPolicy(ctx context.Context, id int16) (exists bool, dp models.DataPolicy, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlDPMGet)
	if err != nil {
		return false, dp, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(
			&dp.Id,
			&dp.Name,
			&dp.Descr,
			&dp.Retention,
			&dp.UseAggr,
			&dp.AggrRetention,
			&dp.AggrInterval,
			&dp.AggrFn,
		)
		if err != nil {
			return false, dp, err
		}
		exists = true
	}
	return exists, dp, err
}

func (pg *PG) GetDataPolicies(ctx context.Context) (dps []models.DataPolicy, err error) {
	dps = []models.DataPolicy{}
	rows, err := pg.db.QueryContext(ctx, sqlDPMGet)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var dp models.DataPolicy
	for rows.Next() {
		err = rows.Scan(
			&dp.Id,
			&dp.Name,
			&dp.Descr,
			&dp.Retention,
			&dp.UseAggr,
			&dp.AggrRetention,
			&dp.AggrInterval,
			&dp.AggrFn,
		)
		if err != nil {
			return nil, err
		}
		dps = append(dps, dp)
	}
	return dps, err
}
