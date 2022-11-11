package db

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/jackc/pgx/v5"
)

type SNMPv2cContainers struct {
	*pgx.Conn
}

type SNMPv2cContainersGetResponse struct {
	// Exists is the container existence.
	Exists bool
	// Container is the SNMPv2c container.
	Container models.SNMPv2cContainer
}

const (
	sqlSNMPv2cContainerGet = `SELECT target, port, cache_duration, transport, community, 
		retries, max_oids, timeout FROM snmpv2c_containers WHERE container_id = $1; `
	sqlSNMPv2cContainerExistsTargetPort = `SELECT EXISTS (SELECT 1 FROM snmpv2c_containers WHERE target = $1 AND port = $2 AND container_id != $3);`
	sqlSNMPv2cContainerCreate           = `INSERT INTO snmpv2c_containers (container_id, target, port, cache_duration, transport, community,
		retries, max_oids, timeout) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);`
	sqlSNMPv2cContainerUpdate = `UPDATE snmpv2c_containers SET (target, port, cache_duration, transport, community,
		retries, max_oids, timeout) = ($1, $2, $3, $4, $5, $6, $7, $8) WHERE container_id = $9;`
)

// Create creates a SNMPv2c container.
func (c *SNMPv2cContainers) Create(ctx context.Context, container models.SNMPv2cContainer) error {
	_, err := c.Exec(ctx, sqlSNMPv2cContainerCreate,
		container.Id,
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

// Update updates a SNMPv2c container.
func (c *SNMPv2cContainers) Update(ctx context.Context, container models.SNMPv2cContainer) (exists bool, err error) {
	t, err := c.Exec(ctx, sqlSNMPv2cContainerUpdate,
		container.Target,
		container.Port,
		container.CacheDuration,
		container.Transport,
		container.Community,
		container.Retries,
		container.MaxOids,
		container.Timeout,
		container.Id,
	)
	return t.RowsAffected() != 0, err
}

// Get returns a SNMPv2c container.
func (c *SNMPv2cContainers) Get(ctx context.Context, containerId int32) (r SNMPv2cContainersGetResponse, err error) {
	rows, err := c.Query(ctx, sqlSNMPv2cContainerGet, containerId)
	if err != nil {
		return r, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(
			&r.Container.Target,
			&r.Container.Port,
			&r.Container.CacheDuration,
			&r.Container.Transport,
			&r.Container.Community,
			&r.Container.Retries,
			&r.Container.MaxOids,
			&r.Container.Timeout,
		)
		if err != nil {
			return r, err
		}
		r.Container.Id = containerId
		r.Exists = true
	}
	return r, nil
}

// AvailableTargetPort returns the existence of an target:port.
func (c *SNMPv2cContainers) AvailableTargetPort(ctx context.Context, target string, port int32, id int32) (exists bool, err error) {
	return exists, c.QueryRow(ctx, sqlSNMPv2cContainerExistsTargetPort, target, port, id).Scan(&exists)
}
