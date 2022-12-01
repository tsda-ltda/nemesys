package pg

import (
	"context"
	"database/sql"

	"github.com/fernandotsda/nemesys/shared/models"
)

const (
	sqlAlarmProfilesCreate = `INSERT INTO alarm_profiles (name, minor, major, critical, emails, wpp, sms, telegrams) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id;`
	sqlAlarmProfilesUpdate = `UPDATE alarm_profiles SET (name, minor, major, critical, emails, wpp, sms, telegrams) = 
		($1, $2, $3, $4, $5, $6, $7, $8) WHERE id = $1;`
	sqlAlarmProfilesGet    = `SELECT name, minor, major, critical, emails, wpp, sms, telegrams FROM alarm_profiles WHERE id = $1;`
	sqlAlarmProfilesMGet   = `SELECT id, name, minor, major, critical, emails, wpp, sms, telegrams FROM alarm_profiles LIMIT $1 OFFSET $2;`
	sqlAlarmProfilesDelete = `DELETE FROM alarm_profiles WHERE id = $1;`
)

func (pg *PG) CreateAlarmProfile(ctx context.Context, profile models.AlarmProfile) (id int64, err error) {
	return id, pg.db.QueryRowContext(ctx, sqlAlarmProfilesCreate,
		profile.Name,
		profile.Minor,
		profile.Major,
		profile.Critical,
		profile.Emails,
		profile.WPP,
		profile.SMS,
		profile.Telegrams,
	).Scan(&id)
}

func (pg *PG) UpdateAlarmProfile(ctx context.Context, profile models.AlarmProfile) (exists bool, err error) {
	t, err := pg.db.ExecContext(ctx, sqlAlarmProfilesUpdate,
		profile.Name,
		profile.Minor,
		profile.Major,
		profile.Critical,
		profile.Emails,
		profile.WPP,
		profile.SMS,
		profile.Telegrams,
		profile.Id,
	)
	if err != nil {
		return false, err
	}
	rowsAffected, _ := t.RowsAffected()
	return rowsAffected != 0, nil
}

func (pg *PG) GetAlarmProfile(ctx context.Context, id int32) (exists bool, profile models.AlarmProfile, err error) {
	err = pg.db.QueryRowContext(ctx, sqlAlarmProfilesGet, id).Scan(
		&profile.Name,
		&profile.Minor,
		&profile.Major,
		&profile.Critical,
		&profile.Emails,
		&profile.WPP,
		&profile.SMS,
		&profile.Telegrams,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, profile, nil
		}
		return false, profile, err
	}
	return true, profile, nil
}

func (pg *PG) GetAlarmProfiles(ctx context.Context, limit int, offset int) (profiles []models.AlarmProfile, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlAlarmProfilesMGet, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	profiles = make([]models.AlarmProfile, 0, limit)
	for rows.Next() {
		var profile models.AlarmProfile
		err = rows.Scan(
			&profile.Id,
			&profile.Name,
			&profile.Minor,
			&profile.Major,
			&profile.Critical,
			&profile.Emails,
			&profile.WPP,
			&profile.SMS,
			&profile.Telegrams,
		)
		if err != nil {
			return nil, err
		}
		profiles = append(profiles, profile)
	}
	return profiles, nil
}

func (pg *PG) DeleteAlarmProfile(ctx context.Context, id int32) (exists bool, err error) {
	return exists, pg.db.QueryRowContext(ctx, sqlAlarmProfilesDelete, id).Scan(&exists)
}
