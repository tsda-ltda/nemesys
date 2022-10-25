package db

import (
	"context"
	"fmt"

	"github.com/fernandotsda/nemesys/shared/env"
	"github.com/jackc/pgx/v5"
)

type PgConn struct {
	*pgx.Conn
	Users          Users
	Teams          Teams
	DataPolicy     DataPolicy
	Contexts       Contexts
	Metrics        Metrics
	Containers     BaseContainers
	SNMPContainers SNMPContainers
	SNMPMetrics    SNMPMetrics
}

// Connects to a Postgresql database server
func ConnectToPG() (*PgConn, error) {
	ctx := context.Background()

	// connect to pg db
	conn, err := pgx.Connect(ctx, fmt.Sprintf("postgres://%s:%s@%s:%s/%s", env.PGUsername, env.PGPW, env.PGHost, env.PGPort, env.PGDBName))
	if err != nil {
		return nil, err
	}

	// ping db
	err = conn.Ping(ctx)
	if err != nil {
		return nil, err
	}

	return &PgConn{
		Conn:           conn,
		Users:          Users{Conn: conn},
		Teams:          Teams{Conn: conn},
		Contexts:       Contexts{Conn: conn},
		Metrics:        Metrics{Conn: conn},
		DataPolicy:     DataPolicy{Conn: conn},
		Containers:     BaseContainers{Conn: conn},
		SNMPContainers: SNMPContainers{Conn: conn},
		SNMPMetrics:    SNMPMetrics{Conn: conn},
	}, nil
}
