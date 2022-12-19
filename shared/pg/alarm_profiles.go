package pg

import (
	"context"
	"database/sql"

	"github.com/fernandotsda/nemesys/shared/models"
)

type AlarmProfileExistsCategoryAndRelationResponse struct {
	// Exists is the alarm profile existence.
	Exists bool
	// CategoryExists is the category existence.
	CategoryExists bool
	// RelationExists is the relation existence.
	RelationExists bool
}

const (
	sqlAlarmProfilesCreate         = `INSERT INTO alarm_profiles (name, descr) VALUES ($1, $2) RETURNING id;`
	sqlAlarmProfilesUpdate         = `UPDATE alarm_profiles SET (name, descr) = ($1, $2) WHERE id = $3;`
	sqlAlarmProfilesGet            = `SELECT name, descr FROM alarm_profiles WHERE id = $1;`
	sqlAlarmProfilesMGet           = `SELECT id, name, descr FROM alarm_profiles LIMIT $1 OFFSET $2;`
	sqlAlarmProfilesDelete         = `DELETE FROM alarm_profiles WHERE id = $1;`
	sqlAlarmProfilesExists         = `SELECT EXISTS (SELECT 1 FROM alarm_profiles WHERE id = $1);`
	sqlAlarmProfilesAddCategory    = `INSERT INTO alarm_profiles_categories_rel (profile_id, category_id) VALUES($1, $2);`
	sqlAlarmProfilesRemoveCategory = `DELETE FROM alarm_profiles_categories_rel WHERE profile_id = $1 AND category_id = $2;`
	sqlAlarmProfilesGetCategories  = `SELECT id, name, descr, level FROM alarm_categories c 
		LEFT JOIN alarm_profiles_categories_rel a ON c.id = a.category_id WHERE a.profile_id = $1 LIMIT $2 OFFSET $3; `
	sqlAlarmProfilesExistsCategoryAndRelation = `SELECT
		EXISTS (SELECT 1 FROM alarm_profiles WHERE id = $1),
		EXISTS (SELECT 1 FROM alarm_categories WHERE id = $2),
		EXISTS (SELECT 1 FROM alarm_profiles_categories_rel WHERE profile_id = $1 AND category_id = $2);`
	sqlAlarmProfilesCreateEmail   = `INSERT INTO alarm_profiles_emails (alarm_profile_id, email) VALUES($1, $2) RETURNING id;`
	sqlAlarmProfilesGetAllEmails  = `SELECT id, email FROM alarm_profiles_emails WHERE alarm_profile_id = $1;`
	sqlAlarmProfilesGetEmails     = `SELECT id, email FROM alarm_profiles_emails WHERE alarm_profile_id = $1 LIMIT $2 OFFSET $3;`
	sqlAlarmProfilesGetOnlyEmails = `SELECT email FROM alarm_profiles_emails WHERE alarm_profile_id = ANY($1);`
	sqlAlarmProfilesDeleteEmail   = `DELETE FROM alarm_profiles_emails WHERE id = $1;`
	sqlAlarmProfilesDeleteEmails  = `DELETE FROM alarm_profiles_emails WHERE alarm_profile_id = $1;`
)

func (pg *PG) CreateAlarmProfile(ctx context.Context, profile models.AlarmProfile) (id int64, err error) {
	return id, pg.db.QueryRowContext(ctx, sqlAlarmProfilesCreate,
		profile.Name,
		profile.Descr,
	).Scan(&id)
}

func (pg *PG) UpdateAlarmProfile(ctx context.Context, profile models.AlarmProfile) (exists bool, err error) {
	t, err := pg.db.ExecContext(ctx, sqlAlarmProfilesUpdate,
		profile.Name,
		profile.Descr,
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
		&profile.Descr,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, profile, nil
		}
		return false, profile, err
	}
	profile.Id = id
	return true, profile, nil
}

func (pg *PG) GetAlarmProfiles(ctx context.Context, limit int, offset int) (profiles []models.AlarmProfile, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlAlarmProfilesMGet, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	profiles = make([]models.AlarmProfile, 0, limit)
	var profile models.AlarmProfile
	for rows.Next() {
		err = rows.Scan(
			&profile.Id,
			&profile.Name,
			&profile.Descr,
		)
		if err != nil {
			return nil, err
		}
		profiles = append(profiles, profile)
	}
	return profiles, nil
}

func (pg *PG) DeleteAlarmProfile(ctx context.Context, id int32) (exists bool, err error) {
	t, err := pg.db.ExecContext(ctx, sqlAlarmProfilesDelete, id)
	if err != nil {
		return false, err
	}
	rowsAffected, _ := t.RowsAffected()
	return rowsAffected != 0, nil
}

func (pg *PG) AlarmProfileExists(ctx context.Context, id int32) (exists bool, err error) {
	return exists, pg.db.QueryRowContext(ctx, sqlAlarmProfilesExists, id).Scan(&exists)
}

func (pg *PG) AddCategoryToAlarmProfile(ctx context.Context, profileId int32, categoryId int32) (err error) {
	_, err = pg.db.ExecContext(ctx, sqlAlarmProfilesAddCategory, profileId, categoryId)
	return err
}

func (pg *PG) RemoveCategoryFromAlarmProfile(ctx context.Context, profileId int32, categoryId int32) (exists bool, err error) {
	t, err := pg.db.ExecContext(ctx, sqlAlarmProfilesRemoveCategory, profileId, categoryId)
	if err != nil {
		return exists, err
	}
	rowsAffected, _ := t.RowsAffected()
	return rowsAffected != 0, nil
}

func (pg *PG) GetAlarmProfileCategories(ctx context.Context, profileId int32, limit int, offset int) (categories []models.AlarmCategory, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlAlarmProfilesGetCategories, profileId, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	categories = make([]models.AlarmCategory, 0)
	var category models.AlarmCategory
	for rows.Next() {
		err = rows.Scan(
			&category.Id,
			&category.Name,
			&category.Descr,
			&category.Level,
		)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	return categories, nil
}

func (pg *PG) CategoryAndAlarmProfileRelationExists(ctx context.Context, profileId int32, categoryId int32) (r AlarmProfileExistsCategoryAndRelationResponse, err error) {
	return r, pg.db.QueryRowContext(ctx, sqlAlarmProfilesExistsCategoryAndRelation, profileId, categoryId).Scan(&r.Exists, &r.CategoryExists, &r.RelationExists)
}

func (pg *PG) CreateAlarmProfileEmail(ctx context.Context, profileId int32, email string) (id int32, err error) {
	return id, pg.db.QueryRowContext(ctx, sqlAlarmProfilesCreateEmail, profileId, email).Scan(&id)
}

func (pg *PG) GetAllAlarmProfileEmails(ctx context.Context, id int32) (emails []models.AlarmProfileEmailWithoutProfileId, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlAlarmProfilesGetAllEmails, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	emails = []models.AlarmProfileEmailWithoutProfileId{}
	var e models.AlarmProfileEmailWithoutProfileId
	for rows.Next() {
		err = rows.Scan(&e.Id, &e.Email)
		if err != nil {
			return nil, err
		}
		emails = append(emails, e)
	}
	return emails, nil
}

func (pg *PG) GetAlarmProfileEmails(ctx context.Context, id int32, limit int, offset int) (emails []models.AlarmProfileEmailWithoutProfileId, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlAlarmProfilesGetEmails, id, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	emails = []models.AlarmProfileEmailWithoutProfileId{}
	var e models.AlarmProfileEmailWithoutProfileId
	for rows.Next() {
		err = rows.Scan(&e.Id, &e.Email)
		if err != nil {
			return nil, err
		}
		emails = append(emails, e)
	}
	return emails, nil
}

func (pg *PG) GetAlarmProfilesEmails(ctx context.Context, ids []int32) (emails []string, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlAlarmProfilesGetOnlyEmails, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	emails = []string{}
	var e string
	for rows.Next() {
		err = rows.Scan(&e)
		if err != nil {
			return nil, err
		}
		emails = append(emails, e)
	}
	return emails, nil
}

func (pg *PG) DeleteAlarmProfileEmail(ctx context.Context, emailId int32) (exists bool, err error) {
	t, err := pg.db.ExecContext(ctx, sqlAlarmProfilesDeleteEmail, emailId)
	if err != nil {
		return exists, err
	}
	rowsAffected, _ := t.RowsAffected()
	return rowsAffected != 0, nil
}

func (pg *PG) DeleteAlarmProfileEmails(ctx context.Context, id int32) (exists bool, err error) {
	t, err := pg.db.ExecContext(ctx, sqlAlarmProfilesDeleteEmails, id)
	if err != nil {
		return false, err
	}
	rowsAffected, _ := t.RowsAffected()
	return rowsAffected != 0, err
}
