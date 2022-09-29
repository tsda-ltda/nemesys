package db

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type TeamsORM struct {
	conn *pgx.Conn
}

// Check by id if team exists
func (u *TeamsORM) Exists(ctx context.Context, id int) (bool, error) {
	var e bool
	sql := `SELECT EXISTS(SELECT 1 FROM teams WHERE id = $1)`
	err := u.conn.QueryRow(ctx, sql, id).Scan(&e)
	return e, err
}
