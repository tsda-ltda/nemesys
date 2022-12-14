package pg

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/models"
)

var ContextValidOrderByColumns = []string{"name", "descr", "ident"}

type ContextQueryFilters struct {
	TeamId    int32  `type:"=" column:"team_id"`
	Name      string `type:"ilike" column:"name"`
	Descr     string `type:"ilike" column:"descr"`
	Ident     string `type:"ilike" column:"ident"`
	OrderBy   string
	OrderByFn string
	Limit     int
	Offset    int
}

func (f ContextQueryFilters) GetOrderBy() string {
	return f.OrderBy
}

func (f ContextQueryFilters) GetOrderByFn() string {
	return f.OrderByFn
}

func (f ContextQueryFilters) GetLimit() int {
	return f.Limit
}

func (f ContextQueryFilters) GetOffset() int {
	return f.Offset
}

// ContextsExistsTeamAndIdentResponse is the response for the Get handler.
type ContextsGetResponse struct {
	// Exists is the context existence.
	Exists bool
	// Contextx is the context.
	Context models.Context
}

// ContextsGetIdsByIdentResponse is the response for the GetIdsByIdent handler.
type ContextsGetIdsByIdentResponse struct {
	// Exists is the context existence.
	Exists bool
	// ContextId is the context id.
	ContextId int32
	// TeamId is the team id.
	TeamId int32
}

const (
	sqlCtxGetIdsByIdent = `WITH tid AS (SELECT id FROM teams WHERE ident = $1)
		SELECT id, (SELECT * FROM tid) FROM contexts WHERE ident = $2 and team_id = (SELECT * FROM tid);`
	sqlCtxExistsTeamAndIdent = `SELECT 
		EXISTS (SELECT 1 FROM teams WHERE id = $1), 
		EXISTS (SELECT 1 FROM contexts WHERE ident = $2 AND team_id = $1 AND id != $3);`
	sqlCtxCreate = `INSERT INTO contexts (ident, descr, name, team_id) VALUES($1, $2, $3, $4) RETURNING id;`
	sqlCtxUpdate = `UPDATE contexts SET (ident, descr, name) = ($1, $2, $3) WHERE id = $4;`
	sqlCtxDelete = `DELETE FROM contexts WHERE id = $1;`
	sqlCtxGet    = `SELECT ident, descr, name, team_id FROM contexts WHERE id = $1;`
	sqlCtxExists = `SELECT EXISTS (SELECT 1 FROM contexts WHERE id = $1);`

	customSqlCtxMGet = `SELECT id, ident, descr, name FROM contexts`
)

func (pg *PG) ContextExists(ctx context.Context, id int32) (exists bool, err error) {
	return exists, pg.db.QueryRowContext(ctx, sqlCtxExists, id).Scan(&exists)
}

func (pg *PG) ExistsTeamAndContextIdent(ctx context.Context, teamId int32, ident string, ctxId int32) (teamExists bool, identExists bool, err error) {
	return teamExists, identExists, pg.db.QueryRowContext(ctx, sqlCtxExistsTeamAndIdent, teamId, ident, ctxId).Scan(
		&teamExists,
		&identExists,
	)
}

func (pg *PG) CreateContext(ctx context.Context, context models.Context) (id int32, err error) {
	return id, pg.db.QueryRowContext(ctx, sqlCtxCreate,
		context.Ident,
		context.Descr,
		context.Name,
		context.TeamId,
	).Scan(&id)
}

func (pg *PG) UpdateContext(ctx context.Context, context models.Context) (exists bool, err error) {
	t, err := pg.db.ExecContext(ctx, sqlCtxCreate,
		context.Ident,
		context.Descr,
		context.Name,
		context.Id,
	)
	if err != nil {
		return false, err
	}
	rowsAffected, _ := t.RowsAffected()
	return rowsAffected != 0, err
}

func (pg *PG) DeleteContext(ctx context.Context, id int32) (exists bool, err error) {
	t, err := pg.db.ExecContext(ctx, sqlCtxDelete, id)
	if err != nil {
		return false, err
	}
	rowsAffected, _ := t.RowsAffected()
	return rowsAffected != 0, err
}

func (pg *PG) GetContexts(ctx context.Context, filters ContextQueryFilters) (contexts []models.Context, err error) {
	sql, params, err := applyFilters(filters, customSqlCtxMGet, ContextValidOrderByColumns)
	if err != nil {
		return nil, err
	}
	rows, err := pg.db.QueryContext(ctx, sql, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	contexts = make([]models.Context, 0, filters.Limit)
	var c models.Context
	for rows.Next() {
		err = rows.Scan(&c.Id, &c.Ident, &c.Descr, &c.Name)
		if err != nil {
			return nil, err
		}
		c.TeamId = filters.TeamId
		contexts = append(contexts, c)
	}
	return contexts, nil
}

func (pg *PG) GetContext(ctx context.Context, id int32) (exists bool, context models.Context, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlCtxGet, id)
	if err != nil {
		return false, context, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(
			&context.Ident,
			&context.Descr,
			&context.Name,
			&context.TeamId,
		)
		if err != nil {
			return false, context, err
		}
		context.Id = id
		exists = true
	}
	return exists, context, nil
}

func (pg *PG) GetContextTreeId(ctx context.Context, ctxIdent string, teamIdent string) (r ContextsGetIdsByIdentResponse, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlCtxGetIdsByIdent, teamIdent, ctxIdent)
	if err != nil {
		return r, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(
			&r.ContextId,
			&r.TeamId,
		)
		if err != nil {
			return r, err
		}
		r.Exists = true
	}
	return r, nil
}
