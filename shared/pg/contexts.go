package pg

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/jackc/pgx/v5"
)

type Contexts struct {
	*pgx.Conn
}

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

// Exists returns the existence of a context.
func (c *Contexts) Exists(ctx context.Context, id int32) (exists bool, err error) {
	return exists, c.QueryRow(ctx, sqlCtxExists, id).Scan(&exists)
}

// ExistsTeamAndIdent returns the existence of a team and a context ident.
func (c *Contexts) ExistsTeamAndIdent(ctx context.Context, teamId int32, ident string, ctxId int32) (r ContextsExistsTeamAndIdentResponse, err error) {
	return r, c.QueryRow(ctx, sqlCtxExistsTeamAndIdent, teamId, ident, ctxId).Scan(
		&r.TeamExists,
		&r.IdentExists,
	)
}

// Create creates a context returning it's id.
func (c *Contexts) Create(ctx context.Context, context models.Context) (id int32, err error) {
	return id, c.QueryRow(ctx, sqlCtxCreate,
		context.Ident,
		context.Descr,
		context.Name,
		context.TeamId,
	).Scan(&id)
}

// Update updates a context.
func (c *Contexts) Update(ctx context.Context, context models.Context) (exists bool, err error) {
	t, err := c.Exec(ctx, sqlCtxCreate,
		context.Ident,
		context.Descr,
		context.Name,
		context.Id,
	)
	return t.RowsAffected() != 0, err
}

// Delete deletes a context.
func (c *Contexts) Delete(ctx context.Context, id int32) (exists bool, err error) {
	t, err := c.Exec(ctx, sqlCtxDelete, id)
	return t.RowsAffected() != 0, err
}

// MGet returns all team's contexts with a limit and a offset.
func (c *Contexts) MGet(ctx context.Context, teamId int32, limit int, offset int) (contexts []models.Context, err error) {
	contexts = []models.Context{}
	rows, err := c.Query(ctx, sqlCtxMGet, teamId, limit, offset)
	if err != nil {
		return contexts, err
	}
	defer rows.Close()
	for rows.Next() {
		var c models.Context
		err = rows.Scan(&c.Id, &c.Ident, &c.Descr, &c.Name)
		if err != nil {
			return contexts, err
		}
		c.TeamId = teamId
		contexts = append(contexts, c)
	}
	return contexts, nil
}

// Get returns a context by id.
func (c *Contexts) Get(ctx context.Context, id int32) (r ContextsGetResponse, err error) {
	rows, err := c.Query(ctx, sqlCtxGet, id)
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

// GetIdsByIdent returns the context and team ids using their ident.
func (c *Contexts) GetIdsByIdent(ctx context.Context, ctxIdent string, teamIdent string) (r ContextsGetIdsByIdentResponse, err error) {
	rows, err := c.Query(ctx, sqlCtxGetIdsByIdent, teamIdent, ctxIdent)
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
