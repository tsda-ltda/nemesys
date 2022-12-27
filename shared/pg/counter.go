package pg

import "context"

const (
	sqlCounterCreate = `INSERT INTO counter_whitelist (user_id) VALUES ($1);`
	sqlCounterMGet   = `SELECT user_id FROM counter_whitelist LIMIT $1 OFFSET $2;`
	sqlCounterDelete = `DELETE FROM counter_whitelist WHERE user_id = $1;`
	sqlCounterGetAll = `SELECT user_id FROM counter_whitelist;`
)

func (pg *PG) AddUserToCounterWhitelist(ctx context.Context, userId int32) (err error) {
	_, err = pg.db.ExecContext(ctx, sqlCounterCreate, userId)
	return err
}

func (pg *PG) GetCounterWhitelist(ctx context.Context, limit int, offset int) (userIds []int32, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlCounterMGet, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	userIds = make([]int32, 0, limit)
	var id int32
	for rows.Next() {
		err = rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		userIds = append(userIds, id)
	}
	return userIds, nil
}

func (pg *PG) RemoveUserFromCounterWhitelist(ctx context.Context, userId int32) (exists bool, err error) {
	t, err := pg.db.ExecContext(ctx, sqlCounterDelete, userId)
	if err != nil {
		return exists, err
	}
	rowsAffected, _ := t.RowsAffected()
	return rowsAffected != 0, nil
}

func (pg *PG) GetAllCounterWhitelist(ctx context.Context) (userIds []int32, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlCounterGetAll)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	userIds = []int32{}
	var id int32
	for rows.Next() {
		err = rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		userIds = append(userIds, id)
	}
	return userIds, nil
}
