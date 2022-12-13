package pg

import (
	"context"
	"database/sql"

	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/types"
)

type FlexLegacyContainersGetProtocolResponse struct {
	// Exists is the flex legacy container existence.
	Exists bool
	// Container is the flex legacy container.
	Container models.FlexLegacyContainer
}

type FlexLegacyContainersGetResponse struct {
	// Exists is the flex legacy container existence.
	Exists bool
	// Container is the flex legacy container.
	Container models.Container[models.FlexLegacyContainer]
}

type FlexLegacyContainersGetSNMPConfigResponse struct {
	// Exists is the flex legacy container existence.
	Exists bool
	// Container is the flex legacy container.
	Container models.FlexLegacyContainer
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
	sqlFlexLegacyContainersGetSNMPConfig = `SELECT
		target, port, transport, community, retries, max_oids, timeout FROM flex_legacy_containers WHERE container_id = $1;`
	sqlFlexLegacyContainersExistsContainerTargetAndSerialNumber = `SELECT
		EXISTS (SELECT 1 FROM containers WHERE id = $1),
		EXISTS (SELECT 1 FROM flex_legacy_containers WHERE target = $2 AND container_id != $1),
		EXISTS (SELECT 1 FROM flex_legacy_containers WHERE serial_number = $3 AND container_id != $1);`
	sqlFlexLegacyContainersGetTarget     = `SELECT target FROM flex_legacy_containers WHERE container_id = $1;`
	sqlFlexLegacyContainersCount         = `SELECT COUNT(*) FROM flex_legacy_containers;`
	sqlFlexLegacyContainersGetIdByTarget = `SELECT container_id FROM flex_legacy_containers WHERE target = $1;`
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
		container.Protocol.Coutry,
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
		container.Protocol.Coutry,
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

func (pg *PG) GetFlexLegacyContainer(ctx context.Context, id int32) (r FlexLegacyContainersGetResponse, err error) {
	baseR, err := pg.GetContainer(ctx, id, types.CTFlexLegacy)
	if err != nil {
		return r, err
	}
	protocolR, err := pg.GetFlexLegacyContainerProtocol(ctx, id)
	if err != nil {
		return r, err
	}
	r.Exists = baseR.Exists
	r.Container.Base = baseR.Container
	r.Container.Protocol = protocolR.Container
	return r, nil
}

func (pg *PG) GetFlexLegacyContainerProtocol(ctx context.Context, id int32) (r FlexLegacyContainersGetProtocolResponse, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlFlexLegacyContainersGetProtocol, id)
	if err != nil {
		return r, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(
			&r.Container.Target,
			&r.Container.Port,
			&r.Container.Transport,
			&r.Container.Community,
			&r.Container.Retries,
			&r.Container.MaxOids,
			&r.Container.Timeout,
			&r.Container.SerialNumber,
			&r.Container.Model,
			&r.Container.City,
			&r.Container.Region,
			&r.Container.Coutry,
		)
		if err != nil {
			return r, err
		}
		r.Exists = true
		r.Container.Id = id
	}
	return r, nil
}

func (pg *PG) GetFlexLegacyContainerSNMPConfig(ctx context.Context, id int32) (r FlexLegacyContainersGetSNMPConfigResponse, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlFlexLegacyContainersGetSNMPConfig, id)
	if err != nil {
		return r, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(
			&r.Container.Target,
			&r.Container.Port,
			&r.Container.Transport,
			&r.Container.Community,
			&r.Container.Retries,
			&r.Container.MaxOids,
			&r.Container.Timeout,
		)
		if err != nil {
			return r, err
		}
		r.Exists = true
		r.Container.Id = id
	}
	return r, nil
}

func (pg *PG) ExistsFlexLegacyContainerTargetPortAndSerialNumber(ctx context.Context, id int32, target string, serialNumber int32) (r FlexLegacyContainerExistsContainerTargetAndSerialNumberRespose, err error) {
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
