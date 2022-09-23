package db

import "github.com/jackc/pgx/v5"

type PgConn struct {
	*pgx.Conn
	Users UsersORM
}
