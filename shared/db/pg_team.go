package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type TeamsORM struct {
	conn *pgx.Conn
}

// Creates the teams table
func (u *TeamsORM) CreateTable(ctx context.Context) (pgconn.CommandTag, error) {
	sql := `CREATE TABLE teams (
		id serial4 PRIMARY KEY,
		name VARCHAR (50) NOT NULL,
		ident VARCHAR (50) UNIQUE NOT NULL,
		descr VARCHAR (255) NOT NULL,
		users_ids INT[] NOT NULL,
	)`
	return u.conn.Exec(ctx, sql)
}

// Check by id if team exists
func (u *TeamsORM) Exists(ctx context.Context, id int) (bool, error) {
	var e bool
	sql := `SELECT EXISTS(SELECT 1 FROM teams WHERE id = $1)`
	err := u.conn.QueryRow(ctx, sql, id).Scan(&e)
	return e, err
}
