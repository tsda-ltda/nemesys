package pg

import (
	"database/sql"
	"fmt"

	"github.com/fernandotsda/nemesys/shared/env"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type PG struct {
	// db is the connection db.
	db *sql.DB
}

func New() *PG {
	url := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", env.PGUsername, env.PGPassword, env.PGHost, env.PGPort, env.PGDBName)
	db, err := sql.Open("pgx", url)
	if err != nil {
		panic("fail to open sql driver, err: " + err.Error())
	}

	db.SetConnMaxIdleTime(0)
	db.SetConnMaxLifetime(0)

	return &PG{db: db}
}

func (pg *PG) Close() {
	pg.db.Close()
}
