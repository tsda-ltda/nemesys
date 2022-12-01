package pg

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/jackc/pgx/v5"
)

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

func (pg *PG) CreateCustomQuery(ctx context.Context, cq models.CustomQuery) (id int32, err error) {

	return id, pg.db.QueryRowContext(ctx, sqlCustomQueriesCreate,
		cq.Ident,
		cq.Descr,
		cq.Flux,
	).Scan(&id)
}

func (pg *PG) UpdateCustomQuery(ctx context.Context, cq models.CustomQuery) (exists bool, err error) {
	t, err := pg.db.ExecContext(ctx, sqlCustomQueriesUpdate,
		cq.Ident,
		cq.Descr,
		cq.Flux,
		cq.Id,
	)
	if err != nil {
		return false, err
	}
	rowsAffected, _ := t.RowsAffected()
	return rowsAffected != 0, err
}

func (pg *PG) GetCustomQueries(ctx context.Context, limit int, offset int) (cqs []models.CustomQuery, err error) {
	cqs = nil
	rows, err := pg.db.QueryContext(ctx, sqlCustomQueriesMGet, limit, offset)
	if err != nil {
		return nil, err
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
			return nil, err
		}
		cqs = append(cqs, cq)
	}
	return cqs, nil
}

func (pg *PG) GetCustomQuery(ctx context.Context, id int32) (r CustomQueriesGetResponse, err error) {
	err = pg.db.QueryRowContext(ctx, sqlCustomQueriesGet, id).Scan(
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

func (pg *PG) GetCustomQueryByIdent(ctx context.Context, ident string) (r CustomQueriesGetResponse, err error) {

	err = pg.db.QueryRowContext(ctx, sqlCustomQueriesGetByIdent, ident).Scan(
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

func (pg *PG) DeleteCustomQuery(ctx context.Context, id int32) (exists bool, err error) {

	t, err := pg.db.ExecContext(ctx, sqlCustomQueriesDelete, id)
	if err != nil {
		return false, err
	}
	if err != nil {
		return false, err
	}
	rowsAffected, _ := t.RowsAffected()
	return rowsAffected != 0, err
}

func (pg *PG) ExistsCustomQueryIdent(ctx context.Context, id int32, ident string) (r CustomQueriesExistsIdent, err error) {

	return r, pg.db.QueryRowContext(ctx, sqlCustomQueriesExistsIdent, id, ident).Scan(&r.IdentExists, &r.Exists)
}
