package db

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// Connects to a Postgresql database server
func ConnectToPG(url string) (*PgConn, error) {
	// connect to pg db
	conn, err := pgx.Connect(context.Background(), url)
	if err != nil {
		return nil, err
	}

	// ping db
	err = conn.Ping(context.Background())
	if err != nil {
		return nil, err
	}

	// create ORMs
	return &PgConn{
		Conn: conn,
		Users: UsersORM{
			conn: conn,
		},
		Teams: TeamsORM{
			conn: conn,
		},
	}, nil
}
