package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type UsersORM struct {
	conn *pgx.Conn
}

// Creates the users table
func (u *UsersORM) CreateTable(ctx context.Context) (pgconn.CommandTag, error) {
	sql := `CREATE TABLE users (
		id serial4 PRIMARY KEY,
		name VARCHAR (50) NOT NULL,
		username VARCHAR (50) UNIQUE NOT NULL,
		password VARCHAR (255) NOT NULL,
		email VARCHAR (255) UNIQUE NOT NULL,
		role INT2 NOT NULL,
		teams_ids INT[]
	)`
	return u.conn.Exec(ctx, sql)
}

// Check by id if user exists
func (u *UsersORM) Exists(ctx context.Context, id int) (bool, error) {
	var e bool
	sql := `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`
	err := u.conn.QueryRow(ctx, sql, id).Scan(&e)
	return e, err
}
