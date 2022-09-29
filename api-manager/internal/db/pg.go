package db

import (
	"context"
	"fmt"

	"github.com/fernandotsda/nemesys/shared/db"
	"github.com/fernandotsda/nemesys/shared/env"
)

// sql commands
const (
	// Creates the users table if not exists.
	sqlCreateUsersTable = `	
		CREATE TABLE IF NOT EXISTS users (
			id serial4 PRIMARY KEY,
			name VARCHAR (50) NOT NULL,
			username VARCHAR (50) UNIQUE NOT NULL,
			password VARCHAR (255) NOT NULL,
			email VARCHAR (255) UNIQUE NOT NULL,
			role INT2 NOT NULL,
			teams_ids INT[]
		);
	`

	// Creates the team table if not exists.
	sqlCreateTeamsTable = `
		CREATE TABLE IF NOT EXISTS teams (
			id serial4 PRIMARY KEY,
			name VARCHAR (50) NOT NULL,
			ident VARCHAR (50) UNIQUE NOT NULL,
			descr VARCHAR (255) NOT NULL,
			users_ids INT[] NOT NULL
		);
	`
)

// PGConnectAndInt connects and initialize Postgresql.
func PGConnectAndInit() (conn *db.PgConn, err error) {
	ctx := context.Background()

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

	return conn, nil
}
