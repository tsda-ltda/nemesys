package initdb

import (
	"context"
	"fmt"

	"github.com/fernandotsda/nemesys/shared/db"
	"github.com/fernandotsda/nemesys/shared/env"
	"github.com/jackc/pgx/v5"
)

// PG creates the database and it's tables if database does not exist.
func PG() (initialized bool, err error) {
	ctx := context.Background()

	// connect to default database
	conn, err := pgx.Connect(ctx, fmt.Sprintf("postgres://%s:%s@%s:%s/postgres", env.PGUsername, env.PGPassword, env.PGHost, env.PGPort))
	if err != nil {
		return false, err
	}
	defer conn.Close(ctx)

	// check if database exists
	sql := `SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname = $1)`
	var exists bool
	err = conn.QueryRow(ctx, sql, env.PGDBName).Scan(&exists)
	if err != nil {
		return false, err
	}
	if exists {
		return false, nil
	}

	// create database
	sql = fmt.Sprintf("CREATE DATABASE %s", env.PGDBName)
	_, err = conn.Exec(ctx, sql)
	if err != nil {
		return false, fmt.Errorf("fail to create database, err: %s", err)
	}

	newConn, err := db.ConnectToPG()
	if err != nil {
		return false, fmt.Errorf("fail to connect to db, err:%s", err)
	}
	defer newConn.Close(ctx)

	// exec commands
	for _, sql := range sqlCommands {
		_, err = newConn.Exec(ctx, sql)
		if err != nil {
			return false, fmt.Errorf("fail to exec command: \"%s\", err: %s", sql, err)
		}
	}

	return true, nil
}
