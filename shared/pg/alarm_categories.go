package pg

import (
	"context"
	"database/sql"

	"github.com/fernandotsda/nemesys/shared/models"
)

var AlarmCategoriesValidOrderByColumns = []string{"descr", "name"}

type AlarmCategoriesQueryFilters struct {
	Name      string `type:"ilike" column:"name"`
	Descr     string `type:"ilike" column:"descr"`
	Level     int32  `type:"=" column:"level"`
	OrderBy   string
	OrderByFn string
	Limit     int
	Offset    int
}

func (f AlarmCategoriesQueryFilters) GetOrderBy() string {
	return f.OrderBy
}

func (f AlarmCategoriesQueryFilters) GetOrderByFn() string {
	return f.OrderByFn
}

func (f AlarmCategoriesQueryFilters) GetLimit() int {
	return f.Limit
}

func (f AlarmCategoriesQueryFilters) GetOffset() int {
	return f.Offset
}

const (
	sqlAlarmCategoriesCreate      = `INSERT INTO alarm_categories (name, descr, level) VALUES($1,$2,$3) RETURNING id;`
	sqlAlarmCategoriesUpdate      = `UPDATE alarm_categories SET (name, descr, level) = ($1,$2,$3) WHERE id = $4;`
	sqlAlarmCategoriesDelete      = `DELETE FROM alarm_categories WHERE id = $1;`
	sqlAlarmCategoriesGet         = `SELECT name, descr, level FROM alarm_categories WHERE id = $1;`
	sqlAlarmCategoriesLevelExists = `SELECT EXISTS (SELECT 1 FROM alarm_categories WHERE level = $1 AND id != $2);`
	sqlAlarmCategoriesExists      = `SELECT EXISTS (SELECT 1 FROM alarm_categories WHERE id = $1);`

	sqlAlarmCategoriesGetProfilesSimplified = `SELECT p.id, p.name FROM alarm_profiles p 
		LEFT JOIN alarm_profiles_categories_rel r ON r.profile_id = p.id WHERE r.category_id = $1;`
	sqlAlarmCategoriesGetSimplifiedByIds    = `SELECT id, level FROM alarm_categories WHERE id = ANY($1) ORDER BY level DESC;`
	sqlAlarmCategoriesCreateTrapIdRel       = `INSERT INTO traps_categories_rel (trap_id, category_id) VALUES ($1, $2);`
	sqlAlarmCategoriesTrapIdRelExists       = `SELECT EXISTS (SELECT 1 FROM traps_categories_rel WHERE trap_id = $1);`
	sqlAlarmCategoriesDeleteTrapIdRel       = `DELETE FROM  traps_categories_rel WHERE trap_id = $1;`
	sqlAlarmCategoriesMGetTrapIdRel         = `SELECT trap_id, category_id FROM traps_categories_rel;`
	sqlAlarmCategoriesGetSimplified         = `SELECT level FROM alarm_categories c WHERE id = $1;`
	sqlAlarmCategoriesGetSimplifiedByTrapId = `SELECT c.id, c.level FROM alarm_categories c
		LEFT JOIN traps_categories_rel r ON r.category_id = c.id WHERE r.trap_id = $1;`
	sqlAlarmCategoriesGetTrapRelByTrapIds = `SELECT trap_id, category_id FROM traps_categories_rel WHERE trap_id = ANY($1);`
	sqlAlarmCategoriesGetTrapRels         = `SELECT trap_id, category_id FROM traps_categories_rel LIMIT $1 OFFSET $2;`

	customSqlAlarmCategoriesMGet = `SELECT id, name, descr, level FROM alarm_categories`
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

func (pg *PG) GetAlarmCategories(ctx context.Context, filters AlarmCategoriesQueryFilters) (categories []models.AlarmCategory, err error) {
	sql, params, err := applyFilters(filters, customSqlAlarmCategoriesMGet, AlarmCategoriesValidOrderByColumns)
	if err != nil {
		return nil, err
	}
	rows, err := pg.db.QueryContext(ctx, sql, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	categories = make([]models.AlarmCategory, 0, filters.Limit)
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
	return categories, err
}

func (pg *PG) CategoryLevelExists(ctx context.Context, level int32, id int32) (exists bool, err error) {
	return exists, pg.db.QueryRowContext(ctx, sqlAlarmCategoriesLevelExists, level, id).Scan(&exists)
}

func (pg *PG) AlarmCategoryExists(ctx context.Context, id int32) (exists bool, err error) {
	return exists, pg.db.QueryRowContext(ctx, sqlAlarmCategoriesExists, id).Scan(&exists)
}

func (pg *PG) GetAlarmCategoriesSimplifiedByIds(ctx context.Context, ids []int32) (categories []models.AlarmCategorySimplified, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlAlarmCategoriesGetSimplifiedByIds, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	categories = make([]models.AlarmCategorySimplified, 0, len(ids))
	var c models.AlarmCategorySimplified
	for rows.Next() {
		err = rows.Scan(&c.Id, &c.Level)
		if err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, nil
}

func (pg *PG) GetCategoryAlarmProfilesSimplified(ctx context.Context, id int32) (profiles []models.AlarmProfileSimplified, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlAlarmCategoriesGetProfilesSimplified, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	profiles = make([]models.AlarmProfileSimplified, 0)
	var p models.AlarmProfileSimplified
	for rows.Next() {
		err = rows.Scan(&p.Id, &p.Name)
		if err != nil {
			return nil, err
		}
		profiles = append(profiles, p)
	}
	return profiles, nil
}

func (pg *PG) CreateTrapCategoryRelation(ctx context.Context, rel models.TrapCategoryRelation) (err error) {
	_, err = pg.db.ExecContext(ctx, sqlAlarmCategoriesCreateTrapIdRel, rel.TrapCategoryId, rel.AlarmCategoryId)
	return err
}

func (pg *PG) TrapCategoryRelationExists(ctx context.Context, trapId int16) (exists bool, err error) {
	return exists, pg.db.QueryRowContext(ctx, sqlAlarmCategoriesTrapIdRelExists, trapId).Scan(&exists)
}

func (pg *PG) DeleteTrapCategoryRelation(ctx context.Context, trapId int16) (exists bool, err error) {
	t, err := pg.db.ExecContext(ctx, sqlAlarmCategoriesDeleteTrapIdRel, trapId)
	if err != nil {
		return false, err
	}
	rowsAffected, _ := t.RowsAffected()
	return rowsAffected != 0, err
}

func (pg *PG) GetTrapCategoryRelations(ctx context.Context) (rels []models.TrapCategoryRelation, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlAlarmCategoriesMGetTrapIdRel)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	rels = []models.TrapCategoryRelation{}
	var rel models.TrapCategoryRelation
	for rows.Next() {
		err = rows.Scan(&rel.TrapCategoryId, &rel.AlarmCategoryId)
		if err != nil {
			return nil, err
		}
		rels = append(rels, rel)
	}
	return rels, nil
}

func (pg *PG) GetAlarmCategorySimplified(ctx context.Context, id int32) (exists bool, category models.AlarmCategorySimplified, err error) {
	err = pg.db.QueryRowContext(ctx, sqlAlarmCategoriesGetSimplified, id).Scan(&category.Level)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, category, nil
		}
		return false, category, err
	}
	category.Id = id
	return true, category, nil
}

func (pg *PG) GetAlarmCategorySimplifiedByTrapId(ctx context.Context, trapId int16) (exists bool, category models.AlarmCategorySimplified, err error) {
	err = pg.db.QueryRowContext(ctx, sqlAlarmCategoriesGetSimplifiedByTrapId, trapId).Scan(&category.Id, &category.Level)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, category, nil
		}
		return false, category, err
	}
	return true, category, nil
}

func (pg *PG) GetTrapCategoriesRelationsByIds(ctx context.Context, trapIds []int16) (rels []models.TrapCategoryRelation, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlAlarmCategoriesGetTrapRelByTrapIds, trapIds)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	rels = make([]models.TrapCategoryRelation, 0, len(trapIds))
	var rel models.TrapCategoryRelation
	for rows.Next() {
		err = rows.Scan(
			&rel.TrapCategoryId,
			&rel.AlarmCategoryId,
		)
		if err != nil {
			return nil, err
		}
		rels = append(rels, rel)
	}
	return rels, nil
}

func (pg *PG) GetTrapCategoriesRelations(ctx context.Context, limit int, offset int) (rels []models.TrapCategoryRelation, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlAlarmCategoriesGetTrapRels, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	rels = make([]models.TrapCategoryRelation, 0, limit)
	var rel models.TrapCategoryRelation
	for rows.Next() {
		err = rows.Scan(
			&rel.TrapCategoryId,
			&rel.AlarmCategoryId,
		)
		if err != nil {
			return nil, err
		}
		rels = append(rels, rel)
	}
	return rels, nil
}
