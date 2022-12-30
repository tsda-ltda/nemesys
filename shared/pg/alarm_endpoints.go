package pg

import (
	"context"
	"database/sql"

	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/vmihailenco/msgpack/v5"
)

var AlarmEndpointValidOrderByColumn = []string{"name, url"}

type AlarmEndpointQueryFilters struct {
	AlarmProfileId int32  `type:"=" column:"alarm_profile_id"`
	Name           string `type:"ilike" column:"name"`
	URL            string `type:"ilike" column:"url"`
	OrderBy        string
	OrderByFn      string
}

func (f AlarmEndpointQueryFilters) GetOrderBy() string {
	return f.OrderBy
}

func (f AlarmEndpointQueryFilters) GetOrderByFn() string {
	return f.OrderByFn
}

type AlarmEndpointRelationExistsRes struct {
	RelationExists      bool
	AlarmProfileExists  bool
	AlarmEndpointExists bool
}

const (
	sqlAlarmEndpointsCreate              = `INSERT INTO alarm_endpoints (name, url, headers) VALUES ($1, $2, $3) RETURNING id;`
	sqlAlarmEndpointsUpdate              = `UPDATE alarm_endpoints SET (name, url, headers) = ($1, $2, $3) WHERE id = $4;`
	sqlAlarmEndpointsDelete              = `DELETE FROM alarm_endpoints WHERE id = $1;`
	sqlAlarmEndpointsGet                 = `SELECT name, url, headers FROM alarm_endpoints WHERE id = $1;`
	sqlAlarmEndpointsAddAlarmProfile     = `INSERT INTO alarm_endpoints_rel (alarm_profile_id, alarm_endpoint_id) VALUES ($1, $2);`
	sqlAlarmEndpointsMGetOfAlarmProfiles = `SELECT id, name, url, headers FROM alarm_endpoints ae
	LEFT JOIN alarm_endpoints_rel aer ON aer.alarm_endpoint_id = ae.id WHERE aer.alarm_profile_id = ANY($1);`
	sqlAlarmEndpointsCreateAlarmProfileRel      = `INSERT INTO alarm_endpoints_rel (alarm_profile_id, alarm_endpoint_id) VALUES($1, $2);`
	sqlAlarmEndpointsDeleteAlarmProfileRel      = `DELETE FROM alarm_endpoints_rel WHERE alarm_profile_id = $1 AND alarm_endpoint_id = $2;`
	sqlAlarmEndpointsAlarmProfileRelationExists = `SELECT 
		EXISTS (SELECT 1 FROM alarm_profiles WHERE id = $1),
		EXISTS (SELECT 1 FROM alarm_endpoints WHERE id = $2),
		EXISTS (SELECT 1 FROM alarm_endpoints_rel WHERE alarm_profile_id = $1 AND alarm_endpoint_id = $2)`

	customSqlAlarmEndpointsMGet               = `SELECT id, name, url, headers FROM alarm_endpoints %s LIMIT $1 OFFSET $2;`
	customSqlAlarmEndpointsMGetOfAlarmProfile = `SELECT id, name, url, headers FROM alarm_endpoints ae
		LEFT JOIN alarm_endpoints_rel aer ON aer.alarm_endpoint_id = ae.id %s LIMIT $1 OFFSET $2;`
)

func (pg *PG) CreateAlarmEndpoint(ctx context.Context, endpoint models.AlarmEndpoint) (id int32, err error) {
	headersByte, err := msgpack.Marshal(endpoint.Headers)
	if err != nil {
		return 0, err
	}

	return id, pg.db.QueryRowContext(ctx, sqlAlarmEndpointsCreate,
		endpoint.Name,
		endpoint.URL,
		headersByte,
	).Scan(&id)
}

func (pg *PG) UpdateAlarmEndpoint(ctx context.Context, endpoint models.AlarmEndpoint) (exists bool, err error) {
	headersByte, err := msgpack.Marshal(endpoint.Headers)
	if err != nil {
		return false, err
	}

	t, err := pg.db.ExecContext(ctx, sqlAlarmEndpointsUpdate,
		endpoint.Name,
		endpoint.URL,
		headersByte,
	)
	if err != nil {
		return false, err
	}
	rowsAffected, _ := t.RowsAffected()
	return rowsAffected != 0, nil
}

func (pg *PG) DeleteAlarmEndpoint(ctx context.Context, id int32) (exists bool, err error) {
	t, err := pg.db.ExecContext(ctx, sqlAlarmEndpointsDelete, id)
	if err != nil {
		return false, err
	}
	rowsAffected, _ := t.RowsAffected()
	return rowsAffected != 0, nil
}

func (pg *PG) GetAlarmEndpoint(ctx context.Context, id int32) (exists bool, endpoint models.AlarmEndpoint, err error) {
	var hbytes []byte
	err = pg.db.QueryRowContext(ctx, sqlAlarmEndpointsGet, id).Scan(
		&endpoint.Name,
		&endpoint.URL,
		&hbytes,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, endpoint, nil
		}
		return false, endpoint, err
	}
	endpoint.Id = id
	err = msgpack.Unmarshal(hbytes, &endpoint.Headers)
	if err != nil {
		return false, endpoint, err
	}
	return true, endpoint, nil
}

func (pg *PG) GetAlarmEndpoints(ctx context.Context, filters AlarmEndpointQueryFilters, limit int, offset int) (endpoints []models.AlarmEndpoint, err error) {
	filters.AlarmProfileId = 0
	sql, err := applyFilters(filters, customSqlAlarmEndpointsMGet, AlarmEndpointValidOrderByColumn)
	if err != nil {
		return nil, err
	}
	rows, err := pg.db.QueryContext(ctx, sql, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	endpoints = make([]models.AlarmEndpoint, 0, limit)
	var endpoint models.AlarmEndpoint
	var hbytes []byte
	for rows.Next() {
		err = rows.Scan(
			&endpoint.Id,
			&endpoint.Name,
			&endpoint.URL,
			&hbytes,
		)
		if err != nil {
			return nil, err
		}
		err = msgpack.Unmarshal(hbytes, &endpoint.Headers)
		if err != nil {
			return nil, err
		}
		endpoints = append(endpoints, endpoint)
	}
	return endpoints, nil
}

func (pg *PG) GetAlamProfileAlarmEndpoints(ctx context.Context, filters AlarmEndpointQueryFilters, limit int, offset int) (endpoints []models.AlarmEndpoint, err error) {
	sql, err := applyFilters(filters, customSqlAlarmEndpointsMGetOfAlarmProfile, AlarmEndpointValidOrderByColumn)
	if err != nil {
		return nil, err
	}
	rows, err := pg.db.QueryContext(ctx, sql, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	endpoints = make([]models.AlarmEndpoint, 0)
	var endpoint models.AlarmEndpoint
	var hbytes []byte
	for rows.Next() {
		err = rows.Scan(
			&endpoint.Id,
			&endpoint.Name,
			&endpoint.URL,
			&hbytes,
		)
		if err != nil {
			return nil, err
		}
		err = msgpack.Unmarshal(hbytes, &endpoint.Headers)
		if err != nil {
			return nil, err
		}
		endpoints = append(endpoints, endpoint)
	}
	return endpoints, nil
}

func (pg *PG) GetAlamProfilesAlarmEndpoints(ctx context.Context, profilesIds []int32) (endpoints []models.AlarmEndpoint, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlAlarmEndpointsMGetOfAlarmProfiles, profilesIds)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	endpoints = make([]models.AlarmEndpoint, 0)
	var endpoint models.AlarmEndpoint
	var hbytes []byte
	for rows.Next() {
		err = rows.Scan(
			&endpoint.Id,
			&endpoint.Name,
			&endpoint.URL,
			&hbytes,
		)
		if err != nil {
			return nil, err
		}
		err = msgpack.Unmarshal(hbytes, &endpoint.Headers)
		if err != nil {
			return nil, err
		}
		endpoints = append(endpoints, endpoint)
	}
	return endpoints, nil
}

func (pg *PG) CreateAlarmEndpointRelation(ctx context.Context, alarmProfileId int32, alarmEndpointId int32) (err error) {
	_, err = pg.db.ExecContext(ctx, sqlAlarmEndpointsCreateAlarmProfileRel, alarmProfileId, alarmEndpointId)
	return err
}

func (pg *PG) DeleteAlarmEndpointRelation(ctx context.Context, alarmProfileId int32, alarmEndpointId int32) (exists bool, err error) {
	t, err := pg.db.ExecContext(ctx, sqlAlarmEndpointsDeleteAlarmProfileRel, alarmProfileId, alarmEndpointId)
	if err != nil {
		return false, err
	}
	rowsAffected, _ := t.RowsAffected()
	return rowsAffected != 0, nil
}

func (pg *PG) AlarmEndpointRelationExists(ctx context.Context, alarmProfileId int32, alarmEndpointId int32) (r AlarmEndpointRelationExistsRes, err error) {
	return r, pg.db.QueryRowContext(ctx, sqlAlarmEndpointsAlarmProfileRelationExists, alarmProfileId, alarmEndpointId).Scan(
		&r.AlarmProfileExists,
		&r.AlarmEndpointExists,
		&r.RelationExists,
	)
}
