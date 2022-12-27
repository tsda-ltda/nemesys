package initdb

import (
	"context"
	"fmt"

	"github.com/fernandotsda/nemesys/shared/env"
	"github.com/jackc/pgx/v5"
)

// PG creates the database and it's tables if database does not exist.
func PG() (initialized bool, err error) {
	ctx := context.Background()

	// connect to default database
	conn, err := connect(ctx, "postgres")
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
	sql = fmt.Sprintf("CREATE DATABASE %s WITH ENCODING 'UTF8'", env.PGDBName)
	_, err = conn.Exec(ctx, sql)
	if err != nil {
		return false, fmt.Errorf("fail to create database, err: %s", err)
	}
	err = conn.Close(ctx)
	if err != nil {
		return false, fmt.Errorf("fail to close connection")
	}

	conn, err = connect(ctx, env.PGDBName)
	if err != nil {
		return false, err
	}
	// exec commands
	for _, sql := range sqlCommands {
		_, err = conn.Exec(ctx, sql)
		if err != nil {
			return false, fmt.Errorf("fail to exec command: \"%s\", err: %s", sql, err)
		}
	}

	return true, nil
}

func connect(ctx context.Context, db string) (*pgx.Conn, error) {
	return pgx.Connect(ctx, fmt.Sprintf("postgres://%s:%s@%s:%s/%s", env.PGUsername, env.PGPassword, env.PGHost, env.PGPort, db))
}
