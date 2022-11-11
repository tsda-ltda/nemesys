package db

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/jackc/pgx/v5"
)

type DataPolicy struct {
	*pgx.Conn
}

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

// Count counts the number of data-policies created.
func (c *DataPolicy) Count(ctx context.Context) (n int64, err error) {
	err = c.QueryRow(ctx, sqlDPCount).Scan(&n)
	return n, err
}

// Create creates a data policy.
func (c *DataPolicy) Create(ctx context.Context, dp models.DataPolicy) (id int16, err error) {
	return id, c.QueryRow(ctx, sqlDPCreate,
		&dp.Descr,
		&dp.UseAggregation,
		&dp.Retention,
		&dp.AggregationRetention,
		&dp.AggregationInterval,
	).Scan(&id)
}

// Update updates a data policy.
func (c *DataPolicy) Update(ctx context.Context, dp models.DataPolicy) (exists bool, err error) {
	t, err := c.Exec(ctx, sqlDPUpdate,
		dp.Descr,
		dp.Retention,
		dp.UseAggregation,
		dp.AggregationRetention,
		dp.AggregationInterval,
	)
	return t.RowsAffected() != 0, err
}

// Delete deletes a data policy.
func (c *DataPolicy) Delete(ctx context.Context, id int16) (exists bool, err error) {
	t, err := c.Exec(ctx, sqlDPDelete, id)
	return t.RowsAffected() != 0, err
}

// Get returns a data policy.
func (c *DataPolicy) Get(ctx context.Context, id int16) (r DataPolicyGetResponse, err error) {
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

// MGet returns all data policies.
func (c *DataPolicy) MGet(ctx context.Context) (dps []models.DataPolicy, err error) {
	dps = []models.DataPolicy{}
	rows, err := c.Query(ctx, sqlDPMGet)
	if err != nil {
		return dps, err
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
			return dps, err
		}
		dps = append(dps, dp)
	}
	return dps, err
}
