package db

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type UsersORM struct {
	conn *pgx.Conn
}

// Check by id if user exists
func (u *UsersORM) Exists(ctx context.Context, id int) (bool, error) {
	var e bool
	sql := `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`
	err := u.conn.QueryRow(ctx, sql, id).Scan(&e)
	return e, err
}
