package db

import (
	"context"
	"fmt"

	"github.com/fernandotsda/nemesys/shared/db"
	"github.com/fernandotsda/nemesys/shared/env"
)

// PGConnectAndInt connects and initialize Postgresql.
func PGConnectAndInit() (conn *db.PgConn, err error) {
	ctx := context.Background()

	// create database
	err = createDatabase(ctx)
	if err != nil {
		return nil, fmt.Errorf("fail to create database, err: %s", err)
	}

	// connect to database
	conn, err = db.ConnectToPG(fmt.Sprintf("postgres://%s:%s@%s:%s/%s", env.PGUsername, env.PGPW, env.PGHost, env.PGPort, env.PGDBName))
	if err != nil {
		return nil, err
	}

	// create users table
	_, err = conn.Exec(ctx, sqlCreateUsersTable)
	if err != nil {
		return nil, fmt.Errorf("fail to create users table, err: %s", err)
	}

	// create teams table
	_, err = conn.Exec(ctx, sqlCreateTeamsTable)
	if err != nil {
		return nil, fmt.Errorf("fail to create teams table, err: %s", err)
	}

	// create teams users realtion table
	_, err = conn.Exec(ctx, sqlCreateUsersTeamsTable)
	if err != nil {
		return nil, fmt.Errorf("fail to create users teams relation table, err: %s", err)
	}

	return conn, nil
}

func createDatabase(ctx context.Context) error {
	// connect to default databae
	conn, err := db.ConnectToPG(fmt.Sprintf("postgres://%s:%s@%s:%s/postgres", env.PGUsername, env.PGPW, env.PGHost, env.PGPort))
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	// check if database exists
	sql := `SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname = $1)`
	var exists bool
	err = conn.QueryRow(ctx, sql, env.PGDBName).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	// create database
	sql = fmt.Sprintf("CREATE DATABASE %s", env.PGDBName)
	_, err = conn.Exec(ctx, sql)
	return err
}
