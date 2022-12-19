package pg

import (
	"context"
	"database/sql"

	"github.com/fernandotsda/nemesys/shared/models"
)

const (
	sqlAPIKeyCreate  = `INSERT INTO apikeys (user_id, created_at, descr, ttl) VALUES ($1, $2, $3, $4) RETURNING id;`
	sqlAPIKeyDelete  = `DELETE FROM apikeys WHERE id = $1 AND user_id = $2;`
	sqlAPIKeyMDelete = `DELETE FROM apikeys WHERE id = ANY($1)`
	sqlAPIKeyMGet    = `SELECT id, ttl, created_at, descr FROM apikeys WHERE user_id = $1;`
)

func (pg *PG) CreateAPIKey(ctx context.Context, apikey models.APIKeyInfo) (id int32, tx *sql.Tx, err error) {
	tx, err = pg.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, nil, err
	}
	err = tx.QueryRowContext(ctx, sqlAPIKeyCreate, apikey.UserId, apikey.CreatedAt, apikey.Descr, apikey.TTL).Scan(&id)
	if err != nil {
		return 0, nil, err
	}
	return id, tx, nil
}

func (pg *PG) DeleteAPIKey(ctx context.Context, id int16, userId int32) (exists bool, tx *sql.Tx, err error) {
	tx, err = pg.db.BeginTx(ctx, nil)
	if err != nil {
		return false, nil, err
	}
	t, err := tx.ExecContext(ctx, sqlAPIKeyDelete, id, userId)
	if err != nil {
		return false, nil, err
	}
	rowsAffected, _ := t.RowsAffected()
	return rowsAffected != 0, tx, nil
}

func (pg *PG) DeleteAPIKeys(ctx context.Context, ids []int16) (err error) {
	_, err = pg.db.ExecContext(ctx, sqlAPIKeyMDelete, ids)
	if err != nil {
		return err
	}
	return nil
}

func (pg *PG) GetAPIKeys(ctx context.Context, userId int32) (keys []models.APIKeyInfo, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlAPIKeyMGet, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	keys = []models.APIKeyInfo{}
	var key models.APIKeyInfo
	for rows.Next() {
		err = rows.Scan(&key.Id, &key.TTL, &key.CreatedAt, &key.Descr)
		if err != nil {
			return nil, err
		}
		key.UserId = userId
		keys = append(keys, key)
	}
	return keys, nil
}
