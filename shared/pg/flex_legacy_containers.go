package pg

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/models"
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

type FlexLegacyContainerExistsContainerTargetPortAndSerialNumberRespose struct {
	// ContainerExists is the container existence.
	ContainerExists bool
	// TargetPortExists is the target port combination existence.
	TargetPortExists bool
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
	sqlFlexLegacyContainersExistsContainerTargetPortAndSerialNumber = `SELECT
		EXISTS (SELECT 1 FROM containers WHERE id = $1),
		EXISTS (SELECT 1 FROM flex_legacy_containers WHERE target = $2 AND port = $3 AND container_id != $1),
		EXISTS (SELECT 1 FROM flex_legacy_containers WHERE serial_number = $4 AND container_id != $1);`
)

func (pg *PG) CreateFlexLegacyContainer(ctx context.Context, container models.Container[models.FlexLegacyContainer]) (err error) {
	c, err := pg.pool.Begin(ctx)
	if err != nil {
		return err
	}
	id, err := pg.createContainer(ctx, c, container.Base)
	if err != nil {
		c.Rollback(ctx)
		return err
	}
	_, err = c.Exec(ctx, sqlFlexLegacyContainersCreate,
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
		c.Rollback(ctx)
		return err
	}

	return c.Commit(ctx)
}

func (pg *PG) UpdateFlexLegacyContainer(ctx context.Context, container models.Container[models.FlexLegacyContainer]) (exists bool, err error) {
	c, err := pg.pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	exists, err = pg.updateContainer(ctx, c, container.Base)
	if err != nil {
		c.Rollback(ctx)
		return false, err
	}
	if !exists {
		return false, nil
	}
	t, err := c.Exec(ctx, sqlFlexLegacyContainersUpdate,
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
		c.Rollback(ctx)
		return false, err
	}
	return t.RowsAffected() != 0, c.Commit(ctx)
}

func (pg *PG) DeleteFlexLegacyContainer(ctx context.Context, id int32) (exists bool, err error) {
	return pg.DeleteContainer(ctx, id)
}

func (pg *PG) GetFlexLegacyContainer(ctx context.Context, id int32) (r FlexLegacyContainersGetResponse, err error) {
	baseR, err := pg.GetContainer(ctx, id)
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
	c, err := pg.pool.Acquire(ctx)
	if err != nil {
		return r, err
	}
	defer c.Release()
	rows, err := c.Query(ctx, sqlFlexLegacyContainersGetProtocol, id)
	if err != nil {
		return r, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(
			&r.Container.Target,
			&r.Container.Port,
			&r.Container.Community,
			&r.Container.Transport,
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
	c, err := pg.pool.Acquire(ctx)
	if err != nil {
		return r, err
	}
	defer c.Release()
	rows, err := c.Query(ctx, sqlFlexLegacyContainersGetSNMPConfig, id)
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

func (pg *PG) ExistsFlexLegacyContainerTargetPortAndSerialNumber(ctx context.Context, id int32, target string, port int32, serialNumber int32) (r FlexLegacyContainerExistsContainerTargetPortAndSerialNumberRespose, err error) {
	c, err := pg.pool.Acquire(ctx)
	if err != nil {
		return r, err
	}
	defer c.Release()
	return r, c.QueryRow(ctx, sqlFlexLegacyContainersExistsContainerTargetPortAndSerialNumber, id, target, port, serialNumber).Scan(
		&r.ContainerExists,
		&r.TargetPortExists,
		&r.SerialNumberExists,
	)
}
