package pg

import (
	"context"
	"database/sql"

	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/types"
)

var SNMPv2cContainerValidOrderByColumns = []string{"name", "descr", "created_at", "target"}

type SNMPv2cContainerQueryFilters struct {
	t              types.ContainerType `type:"=" column:"type"`
	Name           string              `type:"ilike" column:"name"`
	Descr          string              `type:"ilike" column:"descr"`
	CreatedAtStart int64               `type:">=" column:"created_at"`
	CreatedAtStop  int64               `type:"<=" column:"created_at"`
	Enabled        *bool               `type:"=" column:"enabled"`
	Target         string              `type:"ilike" column:"target"`
	OrderBy        string
	OrderByFn      string
}

func (f SNMPv2cContainerQueryFilters) GetOrderBy() string {
	return f.OrderBy
}

func (f SNMPv2cContainerQueryFilters) GetOrderByFn() string {
	return f.OrderByFn
}

func (f SNMPv2cContainerQueryFilters) ContainerType() types.ContainerType {
	return types.CTSNMPv2c
}

const (
	sqlSNMPv2cContainerGet = `SELECT c.name, c.descr, c.enabled, c.rts_pulling_interval, c.created_at,
	p.target, p.port, p.transport, p.community, p.retries, p.max_oids, p.timeout
	FROM containers c FULL JOIN snmpv2c_containers p ON p.container_id = c.id WHERE id = $1;`
	sqlSNMPv2cContainerGetProtocol = `SELECT target, port, transport, community,
		retries, max_oids, timeout FROM snmpv2c_containers WHERE container_id = $1; `
	sqlSNMPv2cContainerExistsTargetPort = `SELECT EXISTS (SELECT 1 FROM snmpv2c_containers WHERE target = $1 AND port = $2 AND container_id != $3);`
	sqlSNMPv2cContainerCreate           = `INSERT INTO snmpv2c_containers (container_id, target, port, transport, community,
		retries, max_oids, timeout) VALUES ($1, $2, $3, $4, $5, $6, $7, $8);`
	sqlSNMPv2cContainerUpdate = `UPDATE snmpv2c_containers SET (target, port, transport, community,
		retries, max_oids, timeout) = ($1, $2, $3, $4, $5, $6, $7) WHERE container_id = $8;`

	customSqlSNMPv2cContainerGet = `SELECT c.id, c.name, c.descr, c.enabled, c.rts_pulling_interval, c.created_at,
		p.target, p.port, p.transport, p.community, p.retries, p.max_oids, p.timeout
		FROM containers c FULL JOIN snmpv2c_containers p ON p.container_id = c.id %s LIMIT $1 OFFSET $2`
)

func (pg *PG) CreateSNMPv2cContainer(ctx context.Context, container models.Container[models.SNMPv2cContainer]) (id int32, err error) {
	c, err := pg.db.BeginTx(ctx, nil)
	if err != nil {
		return id, err
	}
	id, err = pg.createContainer(ctx, c, container.Base)
	if err != nil {
		c.Rollback()
		return id, err
	}
	_, err = c.ExecContext(ctx, sqlSNMPv2cContainerCreate,
		id,
		container.Protocol.Target,
		container.Protocol.Port,
		container.Protocol.Transport,
		container.Protocol.Community,
		container.Protocol.Retries,
		container.Protocol.MaxOids,
		container.Protocol.Timeout,
	)
	if err != nil {
		c.Rollback()
		return id, err
	}
	return id, c.Commit()
}

func (pg *PG) UpdateSNMPv2cContainer(ctx context.Context, container models.Container[models.SNMPv2cContainer]) (exists bool, err error) {
	c, err := pg.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	exists, err = pg.updateContainer(ctx, c, container.Base)
	if err != nil {
		c.Rollback()
		return
	}
	if !exists {
		return false, nil
	}
	t, err := c.ExecContext(ctx, sqlSNMPv2cContainerUpdate,
		container.Protocol.Target,
		container.Protocol.Port,
		container.Protocol.Transport,
		container.Protocol.Community,
		container.Protocol.Retries,
		container.Protocol.MaxOids,
		container.Protocol.Timeout,
		container.Protocol.Id,
	)
	if err != nil {
		c.Rollback()
		return false, err
	}
	rowsAffected, _ := t.RowsAffected()
	return rowsAffected != 0, c.Commit()
}

func (pg *PG) DeleteSNMPv2cContainer(ctx context.Context, id int32) (exists bool, err error) {
	return pg.DeleteContainer(ctx, id)
}

func (pg *PG) GetSNMPv2cContainer(ctx context.Context, id int32) (exists bool, container models.Container[models.SNMPv2cContainer], err error) {
	err = pg.db.QueryRowContext(ctx, sqlSNMPv2cContainerGet, id).Scan(
		&container.Base.Name,
		&container.Base.Descr,
		&container.Base.Enabled,
		&container.Base.RTSPullingInterval,
		&container.Base.CreatedAt,
		&container.Protocol.Target,
		&container.Protocol.Port,
		&container.Protocol.Transport,
		&container.Protocol.Community,
		&container.Protocol.Retries,
		&container.Protocol.MaxOids,
		&container.Protocol.Timeout,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, container, nil
		}
		return false, container, err
	}
	container.Base.Type = types.CTSNMPv2c
	container.Base.Id = id
	container.Protocol.Id = id
	return true, container, nil
}

func (pg *PG) GetSNMPv2cGetContainers(ctx context.Context, filters SNMPv2cContainerQueryFilters, limit int, offset int) (containers []models.Container[models.SNMPv2cContainer], err error) {
	filters.t = types.CTSNMPv2c
	sql, err := applyFilters(filters, customSqlSNMPv2cContainerGet, SNMPv2cContainerValidOrderByColumns)
	if err != nil {
		return nil, err
	}
	rows, err := pg.db.QueryContext(ctx, sql, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	containers = make([]models.Container[models.SNMPv2cContainer], 0, limit)
	container := models.Container[models.SNMPv2cContainer]{}
	for rows.Next() {
		err = rows.Scan(
			&container.Base.Id,
			&container.Base.Name,
			&container.Base.Descr,
			&container.Base.Enabled,
			&container.Base.RTSPullingInterval,
			&container.Base.CreatedAt,
			&container.Protocol.Target,
			&container.Protocol.Port,
			&container.Protocol.Transport,
			&container.Protocol.Community,
			&container.Protocol.Retries,
			&container.Protocol.MaxOids,
			&container.Protocol.Timeout,
		)
		if err != nil {
			return nil, err
		}
		containers = append(containers, container)
	}
	return containers, nil
}

func (pg *PG) GetSNMPv2cContainerProtocol(ctx context.Context, id int32) (exists bool, container models.SNMPv2cContainer, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlSNMPv2cContainerGetProtocol, id)
	if err != nil {
		return false, container, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(
			&container.Target,
			&container.Port,
			&container.Transport,
			&container.Community,
			&container.Retries,
			&container.MaxOids,
			&container.Timeout,
		)
		if err != nil {
			return false, container, err
		}
		container.Id = id
		exists = true
	}
	return exists, container, nil
}

func (pg *PG) AvailableSNMPv2cContainerTargetPort(ctx context.Context, target string, port int32, id int32) (exists bool, err error) {
	return exists, pg.db.QueryRowContext(ctx, sqlSNMPv2cContainerExistsTargetPort, target, port, id).Scan(&exists)
}
