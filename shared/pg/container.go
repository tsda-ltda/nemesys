package pg

import (
	"context"
	"database/sql"

	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/types"
)

type BaseContainerGetResponse struct {
	// Exists is the existence of the base container.
	Exists bool
	// Container is the base container.
	Container models.BaseContainer
}

type BasicContainerGetResponse struct {
	// Exists is the existence of the base container.
	Exists bool
	// Container is the base container.
	Container models.Container[struct{}]
}

type BaseContainerGetRTSConfigResponse struct {
	// Exists is the existence of the base container.
	Exists bool
	// Config is the container RTS config.
	Config models.RTSContainerConfig
}

type BaseContainerEnabledResponse struct {
	// Exists is the existence of the base container.
	Exists bool
	// Enabled is the container enabled status.
	Enabled bool
}

const (
	sqlContainersCreate     = `INSERT INTO containers (name, descr, type, enabled, rts_pulling_interval) VALUES ($1, $2, $3, $4, $5)RETURNING id;`
	sqlContainersGet        = `SELECT name, descr, type, enabled, rts_pulling_interval FROM containers WHERE id = $1 AND type = $2;`
	sqlContainersUpdate     = `UPDATE containers SET (name, descr, enabled, rts_pulling_interval) = ($1, $2, $3, $4) WHERE id = $5;`
	sqlContainersDelete     = `DELETE FROM containers WHERE id = $1;`
	sqlContainersMGet       = `SELECT id, name, descr, enabled, rts_pulling_interval FROM containers WHERE type = $1 LIMIT $2 OFFSET $3;`
	sqlContainersGetRTSInfo = `SELECT rts_pulling_interval FROM containers WHERE id = $1;`
	sqlContainersExists     = `SELECT EXISTS (SELECT 1 FROM containers WHERE id = $1);`
	sqlContainersEnabled    = `SELECT enabled FROM containers WHERE id = $1;`
)

func (pg *PG) CreateBasicContainer(ctx context.Context, container models.Container[struct{}]) (id int32, err error) {
	return id, pg.db.QueryRowContext(ctx, sqlContainersCreate,
		container.Base.Name,
		container.Base.Descr,
		container.Base.Type,
		container.Base.Enabled,
		container.Base.RTSPullingInterval,
	).Scan(&id)
}

func (pg *PG) createContainer(ctx context.Context, tx *sql.Tx, container models.BaseContainer) (id int32, err error) {
	return id, tx.QueryRowContext(ctx, sqlContainersCreate,
		container.Name,
		container.Descr,
		container.Type,
		container.Enabled,
		container.RTSPullingInterval,
	).Scan(&id)
}

func (pg *PG) UpdateBasicContainer(ctx context.Context, container models.Container[struct{}]) (exists bool, err error) {
	t, err := pg.db.ExecContext(ctx, sqlContainersUpdate,
		container.Base.Name,
		container.Base.Descr,
		container.Base.Enabled,
		container.Base.RTSPullingInterval,
		container.Base.Id,
	)
	rowsAffected, _ := t.RowsAffected()
	return rowsAffected != 0, err
}

func (pg *PG) updateContainer(ctx context.Context, tx *sql.Tx, container models.BaseContainer) (exists bool, err error) {
	t, err := tx.ExecContext(ctx, sqlContainersUpdate,
		container.Name,
		container.Descr,
		container.Enabled,
		container.RTSPullingInterval,
		container.Id,
	)
	rowsAffected, _ := t.RowsAffected()
	return rowsAffected != 0, err
}

func (pg *PG) DeleteContainer(ctx context.Context, id int32) (exists bool, err error) {
	t, err := pg.db.ExecContext(ctx, sqlContainersDelete, id)
	rowsAffected, _ := t.RowsAffected()
	return rowsAffected != 0, err
}

func (pg *PG) GetBasicContainer(ctx context.Context, id int32) (r BasicContainerGetResponse, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlContainersGet, id, types.CTBasic)
	if err != nil {
		return r, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(
			&r.Container.Base.Name,
			&r.Container.Base.Descr,
			&r.Container.Base.Type,
			&r.Container.Base.Enabled,
			&r.Container.Base.RTSPullingInterval,
		)
		if err != nil {
			return r, err
		}
		r.Container.Base.Id = id
		r.Exists = true
	}
	return r, nil
}

func (pg *PG) GetContainer(ctx context.Context, id int32) (r BaseContainerGetResponse, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlContainersGet, id)
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

func (pg *PG) GetContainers(ctx context.Context, t types.ContainerType, limit int, offset int) (containers []models.BaseContainer, err error) {
	containers = []models.BaseContainer{}
	rows, err := pg.db.QueryContext(ctx, sqlContainersMGet, t, limit, offset)
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

func (pg *PG) GetContainerRTSConfig(ctx context.Context, id int32) (r BaseContainerGetRTSConfigResponse, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlContainersGetRTSInfo, id)
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

func (pg *PG) ContainerExist(ctx context.Context, id int32) (exists bool, err error) {
	return exists, pg.db.QueryRowContext(ctx, sqlContainersExists, id).Scan(&exists)
}

func (pg *PG) ContainerEnabled(ctx context.Context, id int32) (r BaseContainerEnabledResponse, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlContainersEnabled, id)
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
