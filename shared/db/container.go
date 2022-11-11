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

// BaseContainerGetResponse is the response for the Get handler.
type BaseContainerGetResponse struct {
	// Exists is the existence of the base container.
	Exists bool
	// Container is the base container.
	Container models.BaseContainer
}

// BaseContainerGetRTSConfigResponse is the response for the GetRTSConfig handler.
type BaseContainerGetRTSConfigResponse struct {
	// Exists is the existence of the base container.
	Exists bool
	// Config is the container RTS config.
	Config models.RTSContainerConfig
}

// BaseContainerEnabledResponse is the response for the Enabled handler.
type BaseContainerEnabledResponse struct {
	// Exists is the existence of the base container.
	Exists bool
	// Enabled is the container enabled status.
	Enabled bool
}

const (
	sqlContainersCreate     = `INSERT INTO containers (name, descr, type, enabled, rts_pulling_interval) VALUES ($1, $2, $3, $4, $5)RETURNING id;`
	sqlContainersGet        = `SELECT name, descr, type, enabled, rts_pulling_interval FROM containers WHERE id = $1;`
	sqlContainersUpdate     = `UPDATE containers SET (name, descr, enabled, rts_pulling_interval) = ($1, $2, $3, $4) WHERE id = $5;`
	sqlContainersDelete     = `DELETE FROM containers WHERE id = $1;`
	sqlContainersMGet       = `SELECT id, name, descr, enabled, rts_pulling_interval FROM containers WHERE type = $1 LIMIT $2 OFFSET $3;`
	sqlContainersGetRTSInfo = `SELECT rts_pulling_interval FROM containers WHERE id = $1;`
	sqlContainersExists     = `SELECT EXISTS (SELECT 1 FROM containers WHERE id = $1);`
	sqlContainersEnabled    = `SELECT enabled FROM containers WHERE id = $1;`
)

// Create crates a container returning it's id.
func (c *BaseContainers) Create(ctx context.Context, container models.BaseContainer) (id int, err error) {
	return id, c.QueryRow(ctx, sqlContainersCreate,
		container.Name,
		container.Descr,
		container.Type,
		container.Enabled,
		container.RTSPullingInterval,
	).Scan(&id)
}

// Update updates a container.
func (c *BaseContainers) Update(ctx context.Context, container models.BaseContainer) (exists bool, err error) {
	t, err := c.Exec(ctx, sqlContainersUpdate,
		container.Name,
		container.Descr,
		container.Enabled,
		container.RTSPullingInterval,
		container.Id,
	)
	return t.RowsAffected() != 0, err
}

// Delete deletes a container.
func (c *BaseContainers) Delete(ctx context.Context, id int32) (exists bool, err error) {
	t, err := c.Exec(ctx, sqlContainersDelete, id)
	return t.RowsAffected() != 0, err
}

// Get returns a container by id.
func (c *BaseContainers) Get(ctx context.Context, id int32) (r BaseContainerGetResponse, err error) {
	rows, err := c.Query(ctx, sqlContainersGet, id)
	if err != nil {
		return r, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(
			&r.Container.Name,
			&r.Container.Descr,
			&r.Container.Type,
			&r.Container.Enabled,
			&r.Container.RTSPullingInterval,
		)
		if err != nil {
			return r, err
		}
		r.Container.Id = id
		r.Exists = true
	}
	return r, nil
}

// MGet get all containers of a specific container type with a limit and offset.
func (c *BaseContainers) MGet(ctx context.Context, t types.ContainerType, limit int, offset int) (containers []models.BaseContainer, err error) {
	containers = []models.BaseContainer{}
	rows, err := c.Query(ctx, sqlContainersMGet, t, limit, offset)
	if err != nil {
		return containers, err
	}
	defer rows.Close()
	for rows.Next() {
		var container models.BaseContainer
		err := rows.Scan(
			&container.Id,
			&container.Name,
			&container.Descr,
			&container.Enabled,
			&container.RTSPullingInterval,
		)
		if err != nil {
			return containers, err
		}
		container.Type = t
		containers = append(containers, container)
	}
	return containers, nil
}

// GetRTSConfig returns the RTS configuration of a container.
func (c *BaseContainers) GetRTSConfig(ctx context.Context, id int32) (r BaseContainerGetRTSConfigResponse, err error) {
	rows, err := c.Query(ctx, sqlContainersGetRTSInfo, id)
	if err != nil {
		return r, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&r.Config.PullingInterval)
		if err != nil {
			return r, err
		}
		r.Exists = true
	}
	return r, nil
}

// Exists returns the existence of a container.
func (c *BaseContainers) Exists(ctx context.Context, id int32) (exists bool, err error) {
	return exists, c.QueryRow(ctx, sqlContainersExists, id).Scan(&exists)
}

// Enabled returns the enabled status of a container.
func (c *BaseContainers) Enabled(ctx context.Context, id int32) (r BaseContainerEnabledResponse, err error) {
	rows, err := c.Query(ctx, sqlContainersEnabled, id)
	if err != nil {
		return r, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&r.Enabled)
		if err != nil {
			return r, err
		}
		r.Exists = true
	}
	return r, nil
}
