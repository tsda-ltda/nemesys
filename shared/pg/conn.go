package pg

import (
	"context"
	"strconv"
	"time"

	"github.com/fernandotsda/nemesys/shared/env"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PG struct {
	// pool is the connection pool.
	pool *pgxpool.Pool
}

func New() *PG {
	port, err := strconv.ParseInt(env.PGPort, 0, 64)
	if err != nil {
		panic("fail to parse postgres port, err: " + err.Error())
	}
	pool, err := pgxpool.NewWithConfig(context.Background(), &pgxpool.Config{
		MaxConns:          10,
		MinConns:          1,
		HealthCheckPeriod: time.Second,
		MaxConnIdleTime:   time.Minute * 5,
		ConnConfig: &pgx.ConnConfig{
			Config: pgconn.Config{
				Host:           env.PGHost,
				Port:           uint16(port),
				Database:       env.PGDBName,
				User:           env.PGUsername,
				Password:       env.PGPassword,
				TLSConfig:      nil,
				ConnectTimeout: time.Second * 10,
			},
			Tracer:                   nil,
			StatementCacheCapacity:   20,
			DescriptionCacheCapacity: 20,
		},
	})
	if err != nil {
		panic("fail to create postgres connection pool, err: " + err.Error())
	}
	return &PG{pool: pool}
}

func (pg *PG) Close() {
	pg.pool.Close()
}
