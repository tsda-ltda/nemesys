package pg

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/jackc/pgx/v5"
)

// DataPolicyGetResponse is the response for Get handler.
type DataPolicyGetResponse struct {
	// Exists is the data policy existence.
	Exists bool
	// DataPolicy is the data policy.
	DataPolicy models.DataPolicy
}

const (
	sqlDPCount  = `SELECT COUNT(*) FROM data_policies;`
	sqlDPDelete = `DELETE FROM data_policies WHERE id = $1;`
	sqlDPCreate = `INSERT INTO data_policies 
		(descr, use_aggregation, retention, aggregation_retention, aggregation_interval)
		VALUES ($1, $2, $3, $4, $5) RETURNING id;`
	sqlDPMGet = `SELECT id, descr, retention, use_aggregation, aggregation_retention, aggregation_interval
		FROM data_policies;`
	sqlDPGet = `SELECT id, descr, retention, use_aggregation, aggregation_retention, aggregation_interval
		FROM data_policies WHERE id = $1;`
	sqlDPUpdate = `UPDATE data_policies SET (descr, retention, 
		use_aggregation, aggregation_retention, aggregation_interval) = ($1, $2, $3, $4, $5) WHERE id = $6;`
)

func (pg *PG) CountDataPolicy(ctx context.Context) (n int64, err error) {
	c, err := pg.pool.Acquire(ctx)
	if err != nil {
		return n, err
	}
	defer c.Release()
	err = c.QueryRow(ctx, sqlDPCount).Scan(&n)
	return n, err
}

func (pg *PG) CreateDataPolicy(ctx context.Context, dp models.DataPolicy) (tx pgx.Tx, id int16, err error) {
	c, err := pg.pool.Begin(ctx)
	if err != nil {
		return nil, id, err
	}
	err = c.QueryRow(ctx, sqlDPCreate,
		&dp.Descr,
		&dp.UseAggregation,
		&dp.Retention,
		&dp.AggregationRetention,
		&dp.AggregationInterval,
	).Scan(&id)
	if err != nil {
		return nil, id, err
	}
	return c, id, nil
}

func (pg *PG) UpdateDataPolicy(ctx context.Context, dp models.DataPolicy) (tx pgx.Tx, exists bool, err error) {
	c, err := pg.pool.Begin(ctx)
	if err != nil {
		return nil, false, err
	}
	t, err := c.Exec(ctx, sqlDPUpdate,
		dp.Descr,
		dp.Retention,
		dp.UseAggregation,
		dp.AggregationRetention,
		dp.AggregationInterval,
		dp.Id,
	)
	if err != nil {
		return nil, false, err
	}
	return c, t.RowsAffected() != 0, nil
}

func (pg *PG) DeleteDataPolicy(ctx context.Context, id int16) (exists bool, err error) {
	c, err := pg.pool.Acquire(ctx)
	if err != nil {
		return false, err
	}
	defer c.Release()
	t, err := c.Exec(ctx, sqlDPDelete, id)
	return t.RowsAffected() != 0, err
}

func (pg *PG) GetDataPolicy(ctx context.Context, id int16) (r DataPolicyGetResponse, err error) {
	c, err := pg.pool.Acquire(ctx)
	if err != nil {
		return r, err
	}
	defer c.Release()
	rows, err := c.Query(ctx, sqlDPMGet)
	if err != nil {
		return r, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(
			&r.DataPolicy.Id,
			&r.DataPolicy.Descr,
			&r.DataPolicy.Retention,
			&r.DataPolicy.UseAggregation,
			&r.DataPolicy.AggregationRetention,
			&r.DataPolicy.AggregationInterval,
		)
		if err != nil {
			return r, err
		}
		r.Exists = true
	}
	return r, err
}

func (pg *PG) GetDataPolicies(ctx context.Context) (dps []models.DataPolicy, err error) {
	c, err := pg.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Release()
	dps = []models.DataPolicy{}
	rows, err := c.Query(ctx, sqlDPMGet)
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
			&dp.UseAggregation,
			&dp.AggregationRetention,
			&dp.AggregationInterval,
		)
		if err != nil {
			return nil, err
		}
		dps = append(dps, dp)
	}
	return dps, err
}
