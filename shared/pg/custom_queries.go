package pg

import (
	"context"
	"database/sql"

	"github.com/fernandotsda/nemesys/shared/models"
)

var CustomQueryValidOrderByColumns = []string{"descr", "ident"}

type CustomQueryQueryFilters struct {
	Ident     string `type:"ilike" column:"ident"`
	Descr     string `type:"ilike" column:"descr"`
	OrderBy   string
	OrderByFn string
	Limit     int
	Offset    int
}

func (f CustomQueryQueryFilters) GetOrderBy() string {
	return f.OrderBy
}

func (f CustomQueryQueryFilters) GetOrderByFn() string {
	return f.OrderByFn
}

func (f CustomQueryQueryFilters) GetLimit() int {
	return f.Limit
}

func (f CustomQueryQueryFilters) GetOffset() int {
	return f.Limit
}

const (
	sqlCustomQueriesCreate      = `INSERT INTO custom_queries (ident, descr, flux) VALUES ($1, $2, $3) RETURNING id;`
	sqlCustomQueriesUpdate      = `UPDATE custom_queries SET (ident, descr, flux) = ($1, $2, $3) WHERE id = $4;`
	sqlCustomQueriesGet         = `SELECT ident, descr, flux FROM custom_queries WHERE id = $1;`
	sqlCustomQueriesGetByIdent  = `SELECT id, descr, flux FROM custom_queries WHERE ident = $1;`
	sqlCustomQueriesDelete      = `DELETE FROM custom_queries WHERE id = $1;`
	sqlCustomQueriesExistsIdent = `SELECT 
		EXISTS (SELECT 1 FROM custom_queries WHERE id != $1 AND ident = $2),
		EXISTS (SELECT 1 FROM custom_queries WHERE id = $1);`

	customSqlCustomQueriesMGet = `SELECT id, ident, descr, flux FROM custom_queries`
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

func (pg *PG) GetCustomQueries(ctx context.Context, filters CustomQueryQueryFilters) (cqs []models.CustomQuery, err error) {
	sql, params, err := applyFilters(filters, customSqlCustomQueriesMGet, CustomQueryValidOrderByColumns)
	if err != nil {
		return nil, err
	}
	rows, err := pg.db.QueryContext(ctx, sql, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cqs = make([]models.CustomQuery, 0, filters.Limit)
	var cq models.CustomQuery
	for rows.Next() {
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

func (pg *PG) GetCustomQuery(ctx context.Context, id int32) (exists bool, cq models.CustomQuery, err error) {
	err = pg.db.QueryRowContext(ctx, sqlCustomQueriesGet, id).Scan(
		&cq.Ident,
		&cq.Descr,
		&cq.Flux,
	)
	if err != nil {
		if err != sql.ErrNoRows {
			return false, cq, err
		}
		return false, cq, nil
	}
	cq.Id = id
	return true, cq, nil
}

func (pg *PG) GetCustomQueryByIdent(ctx context.Context, ident string) (exists bool, cq models.CustomQuery, err error) {
	err = pg.db.QueryRowContext(ctx, sqlCustomQueriesGetByIdent, ident).Scan(
		&cq.Id,
		&cq.Descr,
		&cq.Flux,
	)
	if err != nil {
		if err != sql.ErrNoRows {
			return false, cq, err
		}
		return false, cq, nil
	}
	cq.Ident = ident
	return true, cq, nil
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

func (pg *PG) ExistsCustomQueryIdent(ctx context.Context, id int32, ident string) (existsCq bool, identExists bool, err error) {
	return existsCq, identExists, pg.db.QueryRowContext(ctx, sqlCustomQueriesExistsIdent, id, ident).Scan(&identExists, &existsCq)
}
