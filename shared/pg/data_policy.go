package pg

import (
	"context"
	"database/sql"

	"github.com/fernandotsda/nemesys/shared/models"
)

// DataPolicyGetResponse is the response for Get handler.
type DataPolicyGetResponse struct {
	// Exists is the data policy existence.
	Exists bool
	// DataPolicy is the data policy.
	DataPolicy models.DataPolicy
}

const (
	sqlDPCreate = `INSERT INTO data_policies (descr, use_aggr, retention, aggr_retention, aggr_interval, aggr_fn) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id;`
	sqlDPUpdate = `UPDATE data_policies SET (descr, retention, use_aggr, aggr_retention, agg_interval, aggr_fn) = ($1, $2, $3, $4, $5, $6) WHERE id = $7;`
	sqlDPDelete = `DELETE FROM data_policies WHERE id = $1;`
	sqlDPGet    = `SELECT id, descr, retention, use_aggr, aggr_retention, aggr_interval, aggr_fn FROM data_policies WHERE id = $1;`
	sqlDPMGet   = `SELECT id, descr, retention, use_aggr, aggr_retention, aggr_interval, aggr_fn FROM data_policies;`
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
		&dp.Descr,
		&dp.UseAggr,
		&dp.Retention,
		&dp.AggrRetention,
		&dp.AggrInterval,
		&dp.AggrFn,
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
		dp.Descr,
		dp.Retention,
		dp.UseAggr,
		dp.AggrRetention,
		dp.AggrInterval,
		dp.Id,
		dp.AggrFn,
	)
	if err != nil {
		return nil, false, err
	}
	rowsAffected, _ := t.RowsAffected()
	return c, rowsAffected != 0, err
}

func (pg *PG) DeleteDataPolicy(ctx context.Context, id int16) (exists bool, err error) {
	t, err := pg.db.ExecContext(ctx, sqlDPDelete, id)
	rowsAffected, _ := t.RowsAffected()
	return rowsAffected != 0, err
}

func (pg *PG) GetDataPolicy(ctx context.Context, id int16) (r DataPolicyGetResponse, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlDPMGet)
	if err != nil {
		return r, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(
			&r.DataPolicy.Id,
			&r.DataPolicy.Descr,
			&r.DataPolicy.Retention,
			&r.DataPolicy.UseAggr,
			&r.DataPolicy.AggrRetention,
			&r.DataPolicy.AggrInterval,
			&r.DataPolicy.AggrFn,
		)
		if err != nil {
			return r, err
		}
		r.Exists = true
	}
	return r, err
}

func (pg *PG) GetDataPolicies(ctx context.Context) (dps []models.DataPolicy, err error) {
	dps = []models.DataPolicy{}
	rows, err := pg.db.QueryContext(ctx, sqlDPMGet)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var dp models.DataPolicy
		err = rows.Scan(
			&dp.Id,
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
