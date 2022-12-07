package pg

import (
	"context"
	"database/sql"

	"github.com/fernandotsda/nemesys/shared/models"
)

const (
	sqlAlarmCategoriesCreate      = `INSERT INTO alarm_categories (name, descr, level) VALUES($1,$2,$3) RETURNING id;`
	sqlAlarmCategoriesUpdate      = `UPDATE alarm_categories SET (name, descr, level) = ($1,$2,$3) WHERE id = $4;`
	sqlAlarmCategoriesDelete      = `DELETE FROM alarm_categories WHERE id = $1;`
	sqlAlarmCategoriesGet         = `SELECT name, descr, level FROM alarm_categories WHERE id = $1;`
	sqlAlarmCategoriesMGet        = `SELECT id, name, descr, level FROM alarm_categories LIMIT $1 OFFSET $2;`
	sqlAlarmCategoriesLevelExists = `SELECT EXISTS (SELECT 1 FROM alarm_categories WHERE level = $1 AND id != $2);`
	sqlAlarmCategoriesExists      = `SELECT EXISTS (SELECT 1 FROM alarm_categories WHERE id = $1);`
)

func (pg *PG) CreateAlarmCategory(ctx context.Context, category models.AlarmCategory) (id int32, err error) {
	return id, pg.db.QueryRowContext(ctx, sqlAlarmCategoriesCreate, category.Name, category.Descr, category.Level).Scan(&id)
}

func (pg *PG) UpdateAlarmCategory(ctx context.Context, category models.AlarmCategory) (exists bool, err error) {
	t, err := pg.db.ExecContext(ctx, sqlAlarmCategoriesUpdate, category.Name, category.Descr, category.Level, category.Id)
	if err != nil {
		return exists, err
	}
	rowsAffected, _ := t.RowsAffected()
	return rowsAffected != 0, err
}

func (pg *PG) DeleteAlarmCategory(ctx context.Context, id int32) (exists bool, err error) {
	t, err := pg.db.ExecContext(ctx, sqlAlarmCategoriesDelete, id)
	if err != nil {
		return exists, err
	}
	rowsAffected, _ := t.RowsAffected()
	return rowsAffected != 0, err
}

func (pg *PG) GetAlarmCategory(ctx context.Context, id int32) (exists bool, category models.AlarmCategory, err error) {
	err = pg.db.QueryRowContext(ctx, sqlAlarmCategoriesGet, id).Scan(&category.Name, &category.Descr, &category.Level)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, category, nil
		}
		return false, category, err
	}
	category.Id = id
	return true, category, nil
}

func (pg *PG) GetAlarmCategories(ctx context.Context, limit int, offset int) (categories []models.AlarmCategory, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlAlarmCategoriesMGet, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	categories = make([]models.AlarmCategory, 0, limit)
	for rows.Next() {
		var category models.AlarmCategory
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
	return categories, err
}

func (pg *PG) CategoryLevelExists(ctx context.Context, level int32, id int32) (exists bool, err error) {
	return exists, pg.db.QueryRowContext(ctx, sqlAlarmCategoriesLevelExists, level, id).Scan(&exists)
}

func (pg *PG) AlarmCategoryExists(ctx context.Context, id int32) (exists bool, err error) {
	return exists, pg.db.QueryRowContext(ctx, sqlAlarmCategoriesExists, id).Scan(&exists)
}
