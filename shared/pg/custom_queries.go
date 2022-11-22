package pg

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/jackc/pgx/v5"
)

type CustomQueries struct {
	*pgx.Conn
}

// CustomQueriesGetResponse is the response for the Get handler.
type CustomQueriesGetResponse struct {
	// Exists is the custom query existence.
	Exists bool
	// CustomQuery is the custom query.
	CustomQuery models.CustomQuery
}

// CustomQueriesGetByIdentResponse is the response for the GetByIdent handler.
type CustomQueriesGetByIdentResponse struct {
	// Exists is the custom query existence.
	Exists bool
	// CustomQuery is the custom query.
	CustomQuery models.CustomQuery
}

// CustomQueriesExistsIdent is the response for the ExistsIdent handler.
type CustomQueriesExistsIdent struct {
	// Exists is the custom query existence.
	Exists bool
	// CustomQuery is the custom query.
	IdentExists bool
}

const (
	sqlCustomQueriesCreate      = `INSERT INTO custom_queries (ident, descr, flux) VALUES ($1, $2, $3) RETURNING id;`
	sqlCustomQueriesUpdate      = `UPDATE custom_queries SET (ident, descr, flux) = ($1, $2, $3) WHERE id = $4;`
	sqlCustomQueriesMGet        = `SELECT id, ident, descr, flux FROM custom_queries LIMIT $1 OFFSET $2;`
	sqlCustomQueriesGet         = `SELECT ident, descr, flux FROM custom_queries WHERE id = $1;`
	sqlCustomQueriesGetByIdent  = `SELECT id, descr, flux FROM custom_queries WHERE ident = $1;`
	sqlCustomQueriesDelete      = `DELETE FROM custom_queries WHERE id = $1;`
	sqlCustomQueriesExistsIdent = `SELECT 
		EXISTS (SELECT 1 FROM custom_queries WHERE id != $1 AND ident = $2),
		EXISTS (SELECT 1 FROM custom_queries WHERE id = $1);`
)

func (c *CustomQueries) Create(ctx context.Context, cq models.CustomQuery) (id int32, err error) {
	return id, c.QueryRow(ctx, sqlCustomQueriesCreate,
		cq.Ident,
		cq.Descr,
		cq.Flux,
	).Scan(&id)
}

func (c *CustomQueries) Update(ctx context.Context, cq models.CustomQuery) (exists bool, err error) {
	t, err := c.Exec(ctx, sqlCustomQueriesUpdate,
		cq.Ident,
		cq.Descr,
		cq.Flux,
		cq.Id,
	)
	return t.RowsAffected() != 0, err
}

func (c *CustomQueries) MGet(ctx context.Context, limit int, offset int) (cqs []models.CustomQuery, err error) {
	cqs = []models.CustomQuery{}
	rows, err := c.Query(ctx, sqlCustomQueriesMGet, limit, offset)
	if err != nil {
		return cqs, err
	}
	defer rows.Close()
	for rows.Next() {
		var cq models.CustomQuery
		err = rows.Scan(
			&cq.Id,
			&cq.Ident,
			&cq.Descr,
			&cq.Flux,
		)
		if err != nil {
			return cqs, err
		}
		cqs = append(cqs, cq)
	}
	return cqs, nil
}

func (c *CustomQueries) Get(ctx context.Context, id int32) (r CustomQueriesGetResponse, err error) {
	err = c.QueryRow(ctx, sqlCustomQueriesGet, id).Scan(
		&r.CustomQuery.Ident,
		&r.CustomQuery.Descr,
		&r.CustomQuery.Flux,
	)
	if err != nil {
		if err != pgx.ErrNoRows {
			return r, err
		}
		return r, nil
	}
	r.Exists = true
	r.CustomQuery.Id = id
	return r, nil
}

func (c *CustomQueries) GetByIdent(ctx context.Context, ident string) (r CustomQueriesGetResponse, err error) {
	err = c.QueryRow(ctx, sqlCustomQueriesGetByIdent, ident).Scan(
		&r.CustomQuery.Id,
		&r.CustomQuery.Descr,
		&r.CustomQuery.Flux,
	)
	if err != nil {
		if err != pgx.ErrNoRows {
			return r, err
		}
		return r, nil
	}
	r.Exists = true
	r.CustomQuery.Ident = ident
	return r, nil
}

func (c *CustomQueries) Delete(ctx context.Context, id int32) (exists bool, err error) {
	t, err := c.Exec(ctx, sqlCustomQueriesDelete, id)
	if err != nil {
		return false, err
	}
	return t.RowsAffected() != 0, nil
}

func (c *CustomQueries) ExistsIdent(ctx context.Context, id int32, ident string) (r CustomQueriesExistsIdent, err error) {
	return r, c.QueryRow(ctx, sqlCustomQueriesExistsIdent, id, ident).Scan(&r.IdentExists, &r.Exists)
}
