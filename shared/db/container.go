package db

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/types"
	"github.com/jackc/pgx/v5"
)

type BaseContainers struct {
	*pgx.Conn
}

const (
	sqlContainerCreate = `INSERT INTO containers (name, ident, descr, type) VALUES ($1, $2, $3, $4) 
		RETURNING id;`
	sqlContainerGet         = `SELECT name, ident, descr, type FROM containers WHERE id = $1;`
	sqlContainerUpdate      = `UPDATE containers SET (name, ident, descr) = ($1, $2, $3) WHERE id = $4;`
	sqlContainerExistsIdent = `SELECT EXISTS (SELECT 1 FROM containers WHERE ident = $1);`
	sqlContainerDelete      = `DELETE FROM containers WHERE id = $1;`
	sqlContainerMGet        = `SELECT id, name, ident, descr FROM containers WHERE type = $1 LIMIT $2 OFFSET $3;`
)

// Create crates a container. Returns an error if fails to create
func (c *BaseContainers) Create(ctx context.Context, container models.BaseContainer) (id int, err error) {
	err = c.QueryRow(ctx, sqlContainerCreate,
		container.Name,
		container.Ident,
		container.Descr,
		container.Type,
	).Scan(&id)
	return id, err
}

// Update updates a container if exists. Returns an error if fail to update.
func (c *BaseContainers) Update(ctx context.Context, container models.BaseContainer) (e bool, err error) {
	t, err := c.Exec(ctx, sqlContainerUpdate,
		container.Name,
		container.Ident,
		container.Descr,
		container.Id,
	)
	return t.RowsAffected() != 0, err
}

// Delete deletes a container if exists. Returns an error if fails to delete.
func (c *BaseContainers) Delete(ctx context.Context, id int) (e bool, err error) {
	t, err := c.Exec(ctx, sqlContainerDelete, id)
	return t.RowsAffected() != 0, err
}

// Get returns a container by id. Returns an error if fail to get.
func (c *BaseContainers) Get(ctx context.Context, id int) (e bool, container models.BaseContainer, err error) {
	rows, err := c.Query(ctx, sqlContainerGet, id)
	if err != nil {
		return false, container, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(
			&container.Name,
			&container.Ident,
			&container.Descr,
			&container.Type,
		)
		if err != nil {
			return false, container, err
		}
		container.Id = id
	}
	return rows.CommandTag().RowsAffected() != 0, container, nil
}

// MGet get all containers of a container type with a limit and offset.
// Returns an error if fails to get.
func (c *BaseContainers) MGet(ctx context.Context, t types.ContainerType, limit int, offset int) (containers []models.BaseContainer, err error) {
	containers = []models.BaseContainer{}
	rows, err := c.Query(ctx, sqlContainerMGet, t, limit, offset)
	if err != nil {
		return containers, err
	}
	defer rows.Close()
	for rows.Next() {
		var container models.BaseContainer
		err := rows.Scan(
			&container.Id,
			&container.Name,
			&container.Ident,
			&container.Descr,
		)
		if err != nil {
			return containers, err
		}
		container.Type = t
		containers = append(containers, container)
	}
	return containers, nil
}

// ExistsIdent returns the existence of a ident. Returns an error if fails to query.
func (c *BaseContainers) ExistsIdent(ctx context.Context, ident string) (e bool, err error) {
	return e, c.QueryRow(ctx, sqlContainerExistsIdent).Scan(&e)
}
