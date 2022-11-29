package pg

import (
	"database/sql"

	"github.com/fernandotsda/nemesys/shared/models"
	"golang.org/x/net/context"
)

const (
	sqlFlexLegacyDatalogDownloadRegistryCreate = `INSERT INTO flex_legacy_datalog_download_registry 
		(container_id, metering, status, command, virtual) VALUES($1, $2, $3, $4, $5);`
	sqlFlexLegacyDatalogDownloadRegistryUpdate = `UPDATE flex_legacy_datalog_download_registry 
		SET (metering, status, command, virtual) = ($1, $2, $3, $4) WHERE container_id = $5;`
	sqlFlexLegacyDatalogDownloadRegistryGet = `SELECT metering, status, command, virtual 
		FROM flex_legacy_datalog_download_registry WHERE container_id = $1;`
)

func (pg *PG) CreateFlexLegacyDatalogDownloadRegistry(ctx context.Context, r models.FlexLegacyDatalogDownloadRegistry) (err error) {
	_, err = pg.db.ExecContext(ctx, sqlFlexLegacyDatalogDownloadRegistryCreate,
		r.ContainerId,
		r.Metering,
		r.Status,
		r.Command,
		r.Virtual,
	)
	return err
}

func (pg *PG) UpdateFlexLegacyDatalogDownloadRegistry(ctx context.Context, r models.FlexLegacyDatalogDownloadRegistry) (exists bool, err error) {
	t, err := pg.db.ExecContext(ctx, sqlFlexLegacyDatalogDownloadRegistryUpdate,
		r.Metering,
		r.Status,
		r.Command,
		r.Virtual,
		r.ContainerId,
	)
	rowsAffected, _ := t.RowsAffected()
	return rowsAffected != 0, err
}

func (pg *PG) GetFlexLegacyDatalogDownloadRegistry(ctx context.Context, containerId int32) (exists bool, r models.FlexLegacyDatalogDownloadRegistry, err error) {
	err = pg.db.QueryRowContext(ctx, sqlFlexLegacyDatalogDownloadRegistryGet, containerId).Scan(
		&r.Metering,
		&r.Status,
		&r.Command,
		&r.Virtual,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, r, nil
		}
		return false, r, err
	}
	r.ContainerId = containerId
	return true, r, nil
}
