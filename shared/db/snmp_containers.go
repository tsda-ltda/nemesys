package db

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/jackc/pgx/v5"
)

type SNMPContainers struct {
	*pgx.Conn
}

const (
	sqlSNMPContainerGet = `SELECT 
		target, port, cache_duration, transport, community, retries, msg_flag,
		version, max_oids, timeout FROM snmp_containers WHERE container_id = $1; `
	sqlSNMPContainerExistsTargetPort = `SELECT 
		EXISTS (SELECT 1 FROM containers WHERE ident = $1 AND id != $4), 
		EXISTS (SELECT 1 FROM snmp_containers WHERE target = $2 AND port = $3 AND container_id != $4);`
	sqlSNMPContainerCreate = `INSERT INTO snmp_containers (container_id, target, port, cache_duration, transport, community,
		retries, msg_flag, version, max_oids, timeout) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);`
	sqlSNMPContainerUpdate = `UPDATE snmp_containers SET (target, port, cache_duration, transport, community,
		retries, msg_flag, version, max_oids, timeout) = ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) WHERE container_id = $11;`
)

// Create creates a container and a snmp container. Returns an error if fail to create.
func (c *SNMPContainers) Create(ctx context.Context, container models.SNMPContainer) error {
	_, err := c.Exec(ctx, sqlSNMPContainerCreate,
		container.ContainerId,
		container.Target,
		container.Port,
		container.CacheDuration,
		container.Transport,
		container.Community,
		container.Retries,
		container.MsgFlags,
		container.Version,
		container.MaxOids,
		container.Timeout,
	)
	return err
}

// Create creates a container and a snmp container. Returns an error if fail to create.
func (c *SNMPContainers) Update(ctx context.Context, container models.SNMPContainer) (e bool, err error) {
	t, err := c.Exec(ctx, sqlSNMPContainerUpdate,
		container.Target,
		container.Port,
		container.CacheDuration,
		container.Transport,
		container.Community,
		container.Retries,
		container.MsgFlags,
		container.Version,
		container.MaxOids,
		container.Timeout,
		container.ContainerId,
	)
	return t.RowsAffected() != 0, err
}

// Get returns a SNMP Contaier configuration. Returns an error if fails to get.
func (c *SNMPContainers) Get(ctx context.Context, containerId int32) (e bool, conf models.SNMPContainer, err error) {
	rows, err := c.Query(ctx, sqlSNMPContainerGet, containerId)
	if err != nil {
		return false, conf, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(
			&conf.Target,
			&conf.Port,
			&conf.CacheDuration,
			&conf.Transport,
			&conf.Community,
			&conf.Retries,
			&conf.MsgFlags,
			&conf.Version,
			&conf.MaxOids,
			&conf.Timeout,
		)
		if err != nil {
			return false, conf, err
		}
		conf.ContainerId = containerId
		e = true
	}
	return e, conf, nil
}

// ExistsIdentAndTargetPort returns the existence of a ident and target:port. Returns an error if fails to check.
func (c *SNMPContainers) AvailableIdentAndTargetPort(ctx context.Context, ident string, target string, port int32, id int32) (ie bool, tpe bool, err error) {
	rows, err := c.Query(ctx, sqlSNMPContainerExistsTargetPort, ident, target, port, id)
	if err != nil {
		return false, false, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&ie, &tpe)
		if err != nil {
			return false, false, err
		}
	}
	return ie, tpe, nil
}
