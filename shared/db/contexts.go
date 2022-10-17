package db

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/jackc/pgx/v5"
)

type Contexts struct {
	*pgx.Conn
}

const (
	sqlContextsExistsIdent = `SELECT EXISTS (SELECT 1 FROM contexts WHERE ident = $1);`
	sqlContextsCreate      = `INSERT INTO contexts (ident, descr, name, teamId) VALUES($1, $2, $3, $4);`
	sqlContextsDelete      = `DELETE FROM contexts WHERE id = $1;`
	sqlContextsMGet        = `SELECT id, ident, descr, name FROM contexts WHERE teamId = $1 LIMIT $2 OFFSET $3;`
	sqlContextsGet         = `SELECT ident, descr, name, teamId FROM contexts WHERE id = $1;`
)

// ExstsIdent returns the existence of a context's ident.
// Returns an error if fails to check.
func (c *Contexts) ExistsIdent(ctx context.Context, ident string) (e bool, err error) {
	err = c.QueryRow(ctx, sqlContextsExistsIdent, ident).Scan(&e)
	return e, err
}

// Create creates a context. Returns an error if fails to create.
func (c *Contexts) Create(ctx context.Context, context models.Context) error {
	_, err := c.Exec(ctx, sqlContextsCreate,
		context.Ident,
		context.Descr,
		context.Name,
		context.TeamId,
	)
	return err
}

// Delete deletes a context if exists. Returns an error if fails to delete.
func (c *Contexts) Delete(ctx context.Context, id int) (e bool, err error) {
	t, err := c.Exec(ctx, sqlContextsDelete, id)
	return t.RowsAffected() != 0, err
}

// MGet returns all team's contexts with a limit and a offset. Returns an error if fails to get.
func (c *Contexts) MGet(ctx context.Context, teamId int, limit int, offset int) (contexts []models.Context, err error) {
	contexts = []models.Context{}
	rows, err := c.Query(ctx, sqlContextsMGet, teamId, limit, offset)
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

// Get returns a context by id
func (c *Contexts) Get(ctx context.Context, id int) (e bool, context models.Context, err error) {
	rows, err := c.Query(ctx, sqlContextsGet, id)
	if err != nil {
		return false, context, err
	}
	for rows.Next() {
		err = rows.Scan(&context.Ident, &context.Descr, &context.Name, &context.TeamId)
		if err != nil {
			return false, context, err
		}
		context.Id = id
		e = true
	}
	return e, context, nil
}
