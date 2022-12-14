package pg

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/types"
)

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

func (pg *PG) GetSNMPv2cContainer(ctx context.Context, id int32) (exists bool, container models.Container[models.SNMPv2cContainer], err error) {
	exists, base, err := pg.GetContainer(ctx, id, types.CTSNMPv2c)
	if err != nil {
		return false, container, err
	}
	if !exists {
		return false, container, nil
	}
	exists, protocol, err := pg.GetSNMPv2cContainerProtocol(ctx, id)
	if err != nil {
		return false, container, err
	}
	container.Base = base
	container.Protocol = protocol
	return exists, container, nil
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
