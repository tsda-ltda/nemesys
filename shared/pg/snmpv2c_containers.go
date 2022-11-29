package pg

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/models"
)

type SNMPv2cContainersGetResponse struct {
	// Exists is the container existence.
	Exists bool
	// Container is the SNMPv2c container.
	Container models.Container[models.SNMPv2cContainer]
}

type SNMPv2cContainersGetProtocolResponse struct {
	// Exists is the container existence.
	Exists bool
	// Container is the SNMPv2c container.
	Container models.SNMPv2cContainer
}

const (
	sqlSNMPv2cContainerGetProtocol = `SELECT target, port, transport, community,
		retries, max_oids, timeout FROM snmpv2c_containers WHERE container_id = $1; `
	sqlSNMPv2cContainerExistsTargetPort = `SELECT EXISTS (SELECT 1 FROM snmpv2c_containers WHERE target = $1 AND port = $2 AND container_id != $3);`
	sqlSNMPv2cContainerCreate           = `INSERT INTO snmpv2c_containers (container_id, target, port, transport, community,
		retries, max_oids, timeout) VALUES ($1, $2, $3, $4, $5, $6, $7, $8);`
	sqlSNMPv2cContainerUpdate = `UPDATE snmpv2c_containers SET (target, port, transport, community,
		retries, max_oids, timeout) = ($1, $2, $3, $4, $5, $6, $7) WHERE container_id = $8;`
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

func (pg *PG) GetSNMPv2cContainer(ctx context.Context, id int32) (r SNMPv2cContainersGetResponse, err error) {
	baseR, err := pg.GetContainer(ctx, id)
	if err != nil {
		return r, err
	}
	protocolR, err := pg.GetSNMPv2cContainerProtocol(ctx, id)
	if err != nil {
		return r, err
	}
	r.Exists = baseR.Exists
	r.Container.Base = baseR.Container
	r.Container.Protocol = protocolR.Container
	return r, nil
}

func (pg *PG) GetSNMPv2cContainerProtocol(ctx context.Context, id int32) (r SNMPv2cContainersGetProtocolResponse, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlSNMPv2cContainerGetProtocol, id)
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
		r.Container.Id = id
		r.Exists = true
	}
	return r, nil
}

func (pg *PG) AvailableSNMPv2cContainerTargetPort(ctx context.Context, target string, port int32, id int32) (exists bool, err error) {
	return exists, pg.db.QueryRowContext(ctx, sqlSNMPv2cContainerExistsTargetPort, target, port, id).Scan(&exists)
}
