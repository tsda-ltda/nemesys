package db

import (
	"context"
	"fmt"

	"github.com/fernandotsda/nemesys/shared/env"
	"github.com/jackc/pgx/v5"
)

type PgConn struct {
	*pgx.Conn
	Users             Users
	Teams             Teams
	DataPolicy        DataPolicy
	Contexts          Contexts
	ContextualMetrics ContextualMetrics
	Metrics           Metrics
	Containers        BaseContainers
	SNMPv2cContainers SNMPv2cContainers
	SNMPv2cMetrics    SNMPv2cMetrics
}

// Connects to a Postgresql database server
func ConnectToPG() (*PgConn, error) {
	ctx := context.Background()

	// connect to pg db
	conn, err := pgx.Connect(ctx, fmt.Sprintf("postgres://%s:%s@%s:%s/%s", env.PGUsername, env.PGPassword, env.PGHost, env.PGPort, env.PGDBName))
	if err != nil {
		return nil, err
	}

	// ping db
	err = conn.Ping(ctx)
	if err != nil {
		return nil, err
	}

	return &PgConn{
		Conn:              conn,
		Users:             Users{Conn: conn},
		Teams:             Teams{Conn: conn},
		Contexts:          Contexts{Conn: conn},
		ContextualMetrics: ContextualMetrics{Conn: conn},
		Metrics:           Metrics{Conn: conn},
		DataPolicy:        DataPolicy{Conn: conn},
		Containers:        BaseContainers{Conn: conn},
		SNMPv2cContainers: SNMPv2cContainers{Conn: conn},
		SNMPv2cMetrics:    SNMPv2cMetrics{Conn: conn},
	}, nil
}
