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
	conn, err := pgx.Connect(ctx, fmt.Sprintf("postgres://%s:%s@%s:%s/postgres", env.PGUsername, env.PGPW, env.PGHost, env.PGPort))
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

	// create users table
	_, err = newConn.Exec(ctx, sqlCreateUsersTable)
	if err != nil {
		return false, fmt.Errorf("fail to create users table, err: %s", err)
	}

	// create teams table
	_, err = newConn.Exec(ctx, sqlCreateTeamsTable)
	if err != nil {
		return false, fmt.Errorf("fail to create teams table, err: %s", err)
	}

	// create teams users realtion table
	_, err = newConn.Exec(ctx, sqlCreateUsersTeamsTable)
	if err != nil {
		return false, fmt.Errorf("fail to create users teams relation table, err: %s", err)
	}

	// create data-policies table
	_, err = newConn.Exec(ctx, sqlCreateDataPoliciesTable)
	if err != nil {
		return false, fmt.Errorf("fail to create data-policies table, err: %s", err)
	}

	return true, nil
}
