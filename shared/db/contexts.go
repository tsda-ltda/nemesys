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
	sqlContextsCreate      = `INSERT INTO contexts (ident, descr, name, teamId) VALUE($1, $2, $3, $4);`
	sqlContextsDelete      = `DELETE FROM contexts WHERE id = $1;`
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
