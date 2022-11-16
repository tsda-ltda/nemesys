package pg

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/jackc/pgx/v5"
)

type FlexLegacyContainers struct {
	*pgx.Conn
}

// FlexLegacyContainersGetResponse is the response for Get handler.
type FlexLegacyContainersGetResponse struct {
	// Exists is the flex legacy container existence.
	Exists bool
	// Container is the flex legacy container.
	Container models.FlexLegacyContainer
}

// FlexLegacyContainersGetSNMPConfigResponse is the response for GetSNMPConfig handler.
type FlexLegacyContainersGetSNMPConfigResponse struct {
	// Exists is the flex legacy container existence.
	Exists bool
	// Container is the flex legacy container.
	Container models.FlexLegacyContainer
}

// FlexLegacyContainerExistsContainerTargetPortAndSerialNumberRespose is the response for ExistsContainerTargetPortAndSerialNumber handler.
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
		(container_id, target, port, transport, community, retries, max_oids, timeout, cache_duration, serial_number, model, city, region, country) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14);`
	sqlFlexLegacyContainersUpdate = `UPDATE flex_legacy_containers SET 
		(target, port, transport, community, retries, max_oids, timeout, cache_duration, serial_number, model, city, region, country) = 
		($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) WHERE container_id = $14;`
	sqlFlexLegacyContainersGet = `SELECT 
		target, port, transport, community, retries, max_oids, timeout, cache_duration, serial_number, model, city, region, country
		FROM flex_legacy_containers WHERE container_id = $1;`
	sqlFlexLegacyContainersGetSNMPConfig = `SELECT
		target, port, transport, community, retries, max_oids, timeout, cache_duration FROM flex_legacy_containers WHERE container_id = $1;`
	sqlFlexLegacyContainersExistsContainerTargetPortAndSerialNumber = `SELECT
		EXISTS (SELECT 1 FROM containers WHERE id = $1),
		EXISTS (SELECT 1 FROM flex_legacy_containers WHERE target = $2 AND port = $3 AND container_id != $1),
		EXISTS (SELECT 1 FROM flex_legacy_containers WHERE serial_number = $4 AND container_id != $1);`
)

// Create creates a new flex legacy container.
func (c *FlexLegacyContainers) Create(ctx context.Context, container models.FlexLegacyContainer) (err error) {
	_, err = c.Exec(ctx, sqlFlexLegacyContainersCreate,
		container.Id,
		container.Target,
		container.Port,
		container.Transport,
		container.Community,
		container.Retries,
		container.MaxOids,
		container.Timeout,
		container.CacheDuration,
		container.SerialNumber,
		container.Model,
		container.City,
		container.Region,
		container.Coutry,
	)
	return err
}

// Update updates a flex legacy.
func (c *FlexLegacyContainers) Update(ctx context.Context, container models.FlexLegacyContainer) (exists bool, err error) {
	t, err := c.Exec(ctx, sqlFlexLegacyContainersUpdate,
		container.Target,
		container.Port,
		container.Transport,
		container.Community,
		container.Retries,
		container.MaxOids,
		container.Timeout,
		container.CacheDuration,
		container.SerialNumber,
		container.Model,
		container.City,
		container.Region,
		container.Coutry,
		container.Id,
	)
	return t.RowsAffected() != 0, err
}

// Get returns a flex legacy container by id.
func (c *FlexLegacyContainers) Get(ctx context.Context, id int32) (r FlexLegacyContainersGetResponse, err error) {
	rows, err := c.Query(ctx, sqlFlexLegacyContainersGet, id)
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
			&r.Container.CacheDuration,
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

func (c *FlexLegacyContainers) GetSNMPConfig(ctx context.Context, id int32) (r FlexLegacyContainersGetSNMPConfigResponse, err error) {
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
			&r.Container.CacheDuration,
		)
		if err != nil {
			return r, err
		}
		r.Exists = true
		r.Container.Id = id
	}
	return r, nil
}

// ExistsContainerTargetPortAndSerialNumber returns the existence of a container, target port combination and serial-number.
func (c *FlexLegacyContainers) ExistsContainerTargetPortAndSerialNumber(ctx context.Context, id int32, target string, port int32, serialNumber int32) (r FlexLegacyContainerExistsContainerTargetPortAndSerialNumberRespose, err error) {
	return r, c.QueryRow(ctx, sqlFlexLegacyContainersExistsContainerTargetPortAndSerialNumber, id, target, port, serialNumber).Scan(
		&r.ContainerExists,
		&r.TargetPortExists,
		&r.SerialNumberExists,
	)
}
