package pg

import (
	"context"
	"database/sql"
	"time"

	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/types"
)

var BasicContainerValidOrderByColumns = []string{"name", "descr"}

type BasicContainersQueryFilters struct {
	Type           types.ContainerType `type:"=" column:"type"`
	Name           string              `type:"ilike" column:"name"`
	Descr          string              `type:"ilike" column:"descr"`
	CreatedAtStart int64               `type:">=" column:"created_at"`
	CreatedAtStop  int64               `type:"<=" column:"created_at"`
	Enabled        *bool               `type:"=" column:"enabled"`
	OrderBy        string
	OrderByFn      string
	Limit          int
	Offset         int
}

func (f BasicContainersQueryFilters) GetOrderBy() string {
	return f.OrderBy
}

func (f BasicContainersQueryFilters) GetOrderByFn() string {
	return f.OrderByFn
}

func (f BasicContainersQueryFilters) GetLimit() int {
	return f.Limit
}

func (f BasicContainersQueryFilters) GetOffset() int {
	return f.Limit
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
	sqlContainersCreate        = `INSERT INTO containers (name, descr, type, enabled, rts_pulling_interval, created_at) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id;`
	sqlContainersGet           = `SELECT name, descr, enabled, rts_pulling_interval, created_at FROM containers WHERE id = $1 AND type = $2;`
	sqlContainersUpdate        = `UPDATE containers SET (name, descr, enabled, rts_pulling_interval) = ($1, $2, $3, $4) WHERE id = $5;`
	sqlContainersDelete        = `DELETE FROM containers WHERE id = $1;`
	sqlContainersMGet          = `SELECT id, name, descr, enabled, rts_pulling_interval, created_at FROM containers WHERE type = $1 LIMIT $2 OFFSET $3;`
	sqlContainersGetRTSInfo    = `SELECT rts_pulling_interval FROM containers WHERE id = $1;`
	sqlContainersExists        = `SELECT EXISTS (SELECT 1 FROM containers WHERE id = $1);`
	sqlContainersEnabled       = `SELECT enabled FROM containers WHERE id = $1;`
	sqlContainersMGetIdEnabled = `SELECT id FROM containers WHERE enabled = true AND type = $1 LIMIT $2 OFFSET $3;`

	customSqlContainersMGet = `SELECT b.id, b.name, b.descr, b.enabled, b.rts_pulling_interval, b.created_at FROM containers b`
)

func (pg *PG) CreateBasicContainer(ctx context.Context, container models.Container[struct{}]) (id int32, err error) {
	return id, pg.db.QueryRowContext(ctx, sqlContainersCreate,
		container.Base.Name,
		container.Base.Descr,
		container.Base.Type,
		container.Base.Enabled,
		container.Base.RTSPullingInterval,
		time.Now().Unix(),
	).Scan(&id)
}

func (pg *PG) createContainer(ctx context.Context, tx *sql.Tx, container models.BaseContainer) (id int32, err error) {
	return id, tx.QueryRowContext(ctx, sqlContainersCreate,
		container.Name,
		container.Descr,
		container.Type,
		container.Enabled,
		container.RTSPullingInterval,
		time.Now().Unix(),
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
	if err != nil {
		return false, err
	}
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
	if err != nil {
		return false, err
	}
	rowsAffected, _ := t.RowsAffected()
	return rowsAffected != 0, err
}

func (pg *PG) DeleteContainer(ctx context.Context, id int32) (exists bool, err error) {
	t, err := pg.db.ExecContext(ctx, sqlContainersDelete, id)
	if err != nil {
		return false, err
	}
	rowsAffected, _ := t.RowsAffected()
	return rowsAffected != 0, err
}

func (pg *PG) GetBasicContainer(ctx context.Context, id int32) (exists bool, container models.Container[struct{}], err error) {
	exists, c, err := pg.GetContainer(ctx, id, types.CTBasic)
	container.Base = c
	return exists, container, err
}

func (pg *PG) GetContainer(ctx context.Context, id int32, t types.ContainerType) (exists bool, container models.BaseContainer, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlContainersGet, id, t)
	if err != nil {
		return false, container, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(
			&container.Name,
			&container.Descr,
			&container.Enabled,
			&container.RTSPullingInterval,
			&container.CreatedAt,
		)
		if err != nil {
			return false, container, err
		}
		container.Id = id
		container.Type = t
		exists = true
	}
	return exists, container, nil
}

func (pg *PG) GetBasicContainers(ctx context.Context, filters BasicContainersQueryFilters) (containers []models.Container[struct{}], err error) {
	filters.Type = types.CTBasic
	sql, params, err := applyFilters(filters, customSqlContainersMGet, BasicContainerValidOrderByColumns)
	if err != nil {
		return nil, err
	}
	rows, err := pg.db.QueryContext(ctx, sql, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	containers = make([]models.Container[struct{}], 0, filters.Limit)
	container := models.Container[struct{}]{}
	container.Base.Type = types.CTBasic
	for rows.Next() {
		err = rows.Scan(
			&container.Base.Id,
			&container.Base.Name,
			&container.Base.Descr,
			&container.Base.Enabled,
			&container.Base.RTSPullingInterval,
			&container.Base.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		containers = append(containers, container)
	}
	return containers, nil
}

func (pg *PG) GetContainerRTSConfig(ctx context.Context, id int32) (exists bool, config models.RTSContainerConfig, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlContainersGetRTSInfo, id)
	if err != nil {
		return false, config, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&config.PullingInterval)
		if err != nil {
			return false, config, err
		}
		exists = true
	}
	return exists, config, nil
}

func (pg *PG) ContainerExist(ctx context.Context, id int32) (exists bool, err error) {
	return exists, pg.db.QueryRowContext(ctx, sqlContainersExists, id).Scan(&exists)
}

func (pg *PG) ContainerEnabled(ctx context.Context, id int32) (exists bool, enabled bool, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlContainersEnabled, id)
	if err != nil {
		return false, false, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&enabled)
		if err != nil {
			return false, false, err
		}
		exists = true
	}
	return exists, enabled, nil
}

func (pg *PG) GetEnabledContainersIds(ctx context.Context, containerType types.ContainerType, limit int, offset int) (ids []int32, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlContainersMGetIdEnabled, containerType, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int32
		err = rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}
