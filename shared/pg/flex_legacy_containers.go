package pg

import (
	"context"
	"database/sql"

	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/types"
)

var FlexLegacyContainerValidOrderByColumns = []string{"name", "descr", "created_at", "target", "serial_number", "model", "city", "region", "country"}

type FlexLegacyContainerQueryFilters struct {
	Type           types.ContainerType `type:"=" column:"type"`
	Name           string              `type:"ilike" column:"name"`
	Descr          string              `type:"ilike" column:"descr"`
	CreatedAtStart int64               `type:">=" column:"created_at"`
	CreatedAtStop  int64               `type:"<=" column:"created_at"`
	Enabled        *bool               `type:"=" column:"enabled"`
	Target         string              `type:"ilike" column:"target"`
	SerialNumber   string              `type:"=" column:"serial_number"`
	Model          int16               `type:"=" column:"model"`
	City           string              `type:"ilike" column:"city"`
	Region         string              `type:"ilike" column:"region"`
	Country        string              `type:"ilike" column:"country"`
	OrderBy        string
	OrderByFn      string
}

func (f FlexLegacyContainerQueryFilters) GetOrderBy() string {
	return f.OrderBy
}

func (f FlexLegacyContainerQueryFilters) GetOrderByFn() string {
	return f.OrderByFn
}

type FlexLegacyContainerExistsContainerTargetAndSerialNumberRespose struct {
	// ContainerExists is the container existence.
	ContainerExists bool
	// TargetExists is the target existence.
	TargetExists bool
	// SerialNumberExists is the serial-number existence.
	SerialNumberExists bool
}

const (
	sqlFlexLegacyContainersCreate = `INSERT INTO flex_legacy_containers 
		(container_id, target, port, transport, community, retries, max_oids, timeout, serial_number, model, city, region, country) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13);`
	sqlFlexLegacyContainersUpdate = `UPDATE flex_legacy_containers SET 
		(target, port, transport, community, retries, max_oids, timeout, serial_number, model, city, region, country) = 
		($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) WHERE container_id = $13;`
	sqlFlexLegacyContainersGetProtocol = `SELECT 
		target, port, transport, community, retries, max_oids, timeout, serial_number, model, city, region, country
		FROM flex_legacy_containers WHERE container_id = $1;`
	sqlFlexLegacyContainersGet = `SELECT b.name, b.descr, b.enabled, b.rts_pulling_interval, b.created_at,
		p.target, p.port, p.transport, p.community, p.retries, p.max_oids, p.timeout, p.serial_number, p.model, p.city, p.region, p.country
		FROM containers b FULL JOIN flex_legacy_containers p ON p.container_id = b.id WHERE b.id = $1;`
	sqlFlexLegacyContainersGetSNMPConfig = `SELECT
		target, port, transport, community, retries, max_oids, timeout FROM flex_legacy_containers WHERE container_id = $1;`
	sqlFlexLegacyContainersExistsContainerTargetAndSerialNumber = `SELECT
		EXISTS (SELECT 1 FROM containers WHERE id = $1),
		EXISTS (SELECT 1 FROM flex_legacy_containers WHERE target = $2 AND container_id != $1),
		EXISTS (SELECT 1 FROM flex_legacy_containers WHERE serial_number = $3 AND container_id != $1);`
	sqlFlexLegacyContainersGetTarget     = `SELECT target FROM flex_legacy_containers WHERE container_id = $1;`
	sqlFlexLegacyContainersCount         = `SELECT COUNT(*) FROM flex_legacy_containers;`
	sqlFlexLegacyContainersGetIdByTarget = `SELECT container_id FROM flex_legacy_containers WHERE target = $1;`

	customSqlFlexLegacyContainersMGet = `SELECT b.id, b.name, b.descr, b.enabled, b.rts_pulling_interval, b.created_at,
	p.target, p.port, p.transport, p.community, p.retries, p.max_oids, p.timeout, p.serial_number, p.model, p.city, p.region, p.country
	FROM containers b FULL JOIN flex_legacy_containers p ON p.container_id = b.id %s LIMIT $1 OFFSET $2`
)

func (pg *PG) CreateFlexLegacyContainer(ctx context.Context, container models.Container[models.FlexLegacyContainer]) (id int32, err error) {
	c, err := pg.db.BeginTx(ctx, nil)
	if err != nil {
		return id, err
	}
	id, err = pg.createContainer(ctx, c, container.Base)
	if err != nil {
		c.Rollback()
		return id, err
	}
	_, err = c.ExecContext(ctx, sqlFlexLegacyContainersCreate,
		id,
		container.Protocol.Target,
		container.Protocol.Port,
		container.Protocol.Transport,
		container.Protocol.Community,
		container.Protocol.Retries,
		container.Protocol.MaxOids,
		container.Protocol.Timeout,
		container.Protocol.SerialNumber,
		container.Protocol.Model,
		container.Protocol.City,
		container.Protocol.Region,
		container.Protocol.Country,
	)
	if err != nil {
		c.Rollback()
		return id, err
	}

	return id, c.Commit()
}

func (pg *PG) UpdateFlexLegacyContainer(ctx context.Context, container models.Container[models.FlexLegacyContainer]) (exists bool, err error) {
	c, err := pg.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	exists, err = pg.updateContainer(ctx, c, container.Base)
	if err != nil {
		c.Rollback()
		return false, err
	}
	if !exists {
		return false, nil
	}
	t, err := c.ExecContext(ctx, sqlFlexLegacyContainersUpdate,
		container.Protocol.Target,
		container.Protocol.Port,
		container.Protocol.Transport,
		container.Protocol.Community,
		container.Protocol.Retries,
		container.Protocol.MaxOids,
		container.Protocol.Timeout,
		container.Protocol.SerialNumber,
		container.Protocol.Model,
		container.Protocol.City,
		container.Protocol.Region,
		container.Protocol.Country,
		container.Protocol.Id,
	)
	if err != nil {
		c.Rollback()
		return false, err
	}
	rowsAffected, _ := t.RowsAffected()
	return rowsAffected != 0, c.Commit()
}

func (pg *PG) DeleteFlexLegacyContainer(ctx context.Context, id int32) (exists bool, err error) {
	return pg.DeleteContainer(ctx, id)
}

func (pg *PG) GetFlexLegacyContainers(ctx context.Context, filters FlexLegacyContainerQueryFilters, limit int, offset int) (containers []models.Container[models.FlexLegacyContainer], err error) {
	filters.Type = types.CTFlexLegacy
	sql, err := applyFilters(filters, customSqlFlexLegacyContainersMGet, FlexLegacyContainerValidOrderByColumns)
	if err != nil {
		return nil, err
	}
	rows, err := pg.db.QueryContext(ctx, sql, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	containers = make([]models.Container[models.FlexLegacyContainer], 0, limit)
	container := models.Container[models.FlexLegacyContainer]{}
	container.Base.Type = filters.Type
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
			&container.Protocol.SerialNumber,
			&container.Protocol.Model,
			&container.Protocol.City,
			&container.Protocol.Region,
			&container.Protocol.Country,
		)
		if err != nil {
			return nil, err
		}
		container.Protocol.Id = container.Base.Id
		containers = append(containers, container)
	}
	return containers, nil
}

func (pg *PG) GetFlexLegacyContainer(ctx context.Context, id int32) (exists bool, container models.Container[models.FlexLegacyContainer], err error) {
	err = pg.db.QueryRowContext(ctx, sqlFlexLegacyContainersGet, id).Scan(
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
		&container.Protocol.SerialNumber,
		&container.Protocol.Model,
		&container.Protocol.City,
		&container.Protocol.Region,
		&container.Protocol.Country,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, container, nil
		}
		return false, container, err
	}
	container.Base.Type = types.CTFlexLegacy
	container.Base.Id = id
	container.Protocol.Id = id
	return true, container, nil
}

func (pg *PG) GetFlexLegacyContainerProtocol(ctx context.Context, id int32) (exists bool, container models.FlexLegacyContainer, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlFlexLegacyContainersGetProtocol, id)
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
			&container.SerialNumber,
			&container.Model,
			&container.City,
			&container.Region,
			&container.Country,
		)
		if err != nil {
			return false, container, err
		}
		exists = true
		container.Id = id
	}
	return exists, container, nil
}

func (pg *PG) GetFlexLegacyContainerSNMPConfig(ctx context.Context, id int32) (exists bool, container models.FlexLegacyContainer, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlFlexLegacyContainersGetSNMPConfig, id)
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
		exists = true
		container.Id = id
	}
	return exists, container, nil
}

func (pg *PG) ExistsFlexLegacyContainerTargetPortAndSerialNumber(ctx context.Context, id int32, target string, serialNumber string) (r FlexLegacyContainerExistsContainerTargetAndSerialNumberRespose, err error) {
	return r, pg.db.QueryRowContext(ctx, sqlFlexLegacyContainersExistsContainerTargetAndSerialNumber, id, target, serialNumber).Scan(
		&r.ContainerExists,
		&r.TargetExists,
		&r.SerialNumberExists,
	)
}

func (pg *PG) GetFlexLegacyContainerTarget(ctx context.Context, id int32) (exists bool, target string, err error) {
	err = pg.db.QueryRowContext(ctx, sqlFlexLegacyContainersGetTarget, id).Scan(&target)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, target, nil
		}
		return false, target, err
	}
	return true, target, nil
}

func (pg *PG) CountFlexLegacyContainers(ctx context.Context) (n int, err error) {
	return n, pg.db.QueryRowContext(ctx, sqlFlexLegacyContainersCount).Scan(&n)
}

func (pg *PG) GetFlexLegacyContainerIdByTargetPort(ctx context.Context, target string) (exists bool, id int32, err error) {
	err = pg.db.QueryRowContext(ctx, sqlFlexLegacyContainersGetIdByTarget, target).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, id, nil
		}
		return false, id, err
	}
	return true, id, nil
}
