package db

import "github.com/jackc/pgx/v5"

type PgConn struct {
	*pgx.Conn
	Users      Users
	Teams      Teams
	DataPolicy DataPolicy
	Contexts   Contexts
}
