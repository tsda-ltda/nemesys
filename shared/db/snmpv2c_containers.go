package db

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/jackc/pgx/v5"
)

type SNMPv2cContainers struct {
	*pgx.Conn
}

const (
	sqlSNMPContainerGet = `SELECT target, port, cache_duration, transport, community, 
		retries, max_oids, timeout FROM snmpv2c_containers WHERE container_id = $1; `
	sqlSNMPContainerExistsTargetPort = `SELECT EXISTS (SELECT 1 FROM snmpv2c_containers WHERE target = $1 AND port = $2 AND container_id != $3);`
	sqlSNMPContainerCreate           = `INSERT INTO snmpv2c_containers (container_id, target, port, cache_duration, transport, community,
		retries, max_oids, timeout) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);`
	sqlSNMPContainerUpdate = `UPDATE snmpv2c_containers SET (target, port, cache_duration, transport, community,
		retries, max_oids, timeout) = ($1, $2, $3, $4, $5, $6, $7, $8) WHERE container_id = $9;`
)

// Create creates a container and a snmp container. Returns an error if fail to create.
func (c *SNMPv2cContainers) Create(ctx context.Context, container models.SNMPv2cContainer) error {
	_, err := c.Exec(ctx, sqlSNMPContainerCreate,
		container.ContainerId,
		container.Target,
		container.Port,
		container.CacheDuration,
		container.Transport,
		container.Community,
		container.Retries,
		container.MaxOids,
		container.Timeout,
	)
	return err
}

// Create creates a container and a snmp container. Returns an error if fail to create.
func (c *SNMPv2cContainers) Update(ctx context.Context, container models.SNMPv2cContainer) (e bool, err error) {
	t, err := c.Exec(ctx, sqlSNMPContainerUpdate,
		container.Target,
		container.Port,
		container.CacheDuration,
		container.Transport,
		container.Community,
		container.Retries,
		container.MaxOids,
		container.Timeout,
		container.ContainerId,
	)
	return t.RowsAffected() != 0, err
}

// Get returns a SNMP Contaier configuration. Returns an error if fails to get.
func (c *SNMPv2cContainers) Get(ctx context.Context, containerId int32) (e bool, conf models.SNMPv2cContainer, err error) {
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

// AvailableTargetPort returns the existence of an target:port. Returns an error if fails to check.
func (c *SNMPv2cContainers) AvailableTargetPort(ctx context.Context, target string, port int32, id int32) (tpe bool, err error) {
	rows, err := c.Query(ctx, sqlSNMPContainerExistsTargetPort, target, port, id)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&tpe)
		if err != nil {
			return false, err
		}
	}
	return tpe, nil
}
