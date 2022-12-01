package pg

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/models"
)

// ContextsExistsTeamAndIdentResponse is the response for the ExistsTeamAndIdentResponse handler.
type ContextsExistsTeamAndIdentResponse struct {
	// TeamExists is the team existence.
	TeamExists bool
	// IdentExsts is the ident existence.
	IdentExists bool
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
		EXISTS (SELECT 1 FROM contexts WHERE ident = $2 AND id != $3);`
	sqlCtxCreate = `INSERT INTO contexts (ident, descr, name, team_id) VALUES($1, $2, $3, $4) RETURNING id;`
	sqlCtxUpdate = `UPDATE contexts SET (ident, descr, name) = ($1, $2, $3) WHERE id = $4;`
	sqlCtxDelete = `DELETE FROM contexts WHERE id = $1;`
	sqlCtxMGet   = `SELECT id, ident, descr, name FROM contexts WHERE team_id = $1 LIMIT $2 OFFSET $3;`
	sqlCtxGet    = `SELECT ident, descr, name, team_id FROM contexts WHERE id = $1;`
	sqlCtxExists = `SELECT EXISTS (SELECT 1 FROM contexts WHERE id = $1);`
)

func (pg *PG) ContextExists(ctx context.Context, id int32) (exists bool, err error) {
	return exists, pg.db.QueryRowContext(ctx, sqlCtxExists, id).Scan(&exists)
}

func (pg *PG) ExistsTeamAndContextIdent(ctx context.Context, teamId int32, ident string, ctxId int32) (r ContextsExistsTeamAndIdentResponse, err error) {
	return r, pg.db.QueryRowContext(ctx, sqlCtxExistsTeamAndIdent, teamId, ident, ctxId).Scan(
		&r.TeamExists,
		&r.IdentExists,
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

func (pg *PG) GetContexts(ctx context.Context, teamId int32, limit int, offset int) (contexts []models.Context, err error) {
	contexts = []models.Context{}
	rows, err := pg.db.QueryContext(ctx, sqlCtxMGet, teamId, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var c models.Context
		err = rows.Scan(&c.Id, &c.Ident, &c.Descr, &c.Name)
		if err != nil {
			return nil, err
		}
		c.TeamId = teamId
		contexts = append(contexts, c)
	}
	return contexts, nil
}

func (pg *PG) GetContext(ctx context.Context, id int32) (r ContextsGetResponse, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlCtxGet, id)
	if err != nil {
		return r, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(
			&r.Context.Ident,
			&r.Context.Descr,
			&r.Context.Name,
			&r.Context.TeamId,
		)
		if err != nil {
			return r, err
		}
		r.Context.Id = id
		r.Exists = true
	}
	return r, nil
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
