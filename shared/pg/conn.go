package pg

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

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

	maxConn, err1 := strconv.Atoi(env.PGMaxOpenConn)
	maxIdle, err2 := strconv.Atoi(env.PGMaxIdleConn)
	idleLifetime, err3 := strconv.Atoi(env.PGMaxIdleConnLifetime)
	lifetime, err4 := strconv.Atoi(env.PGMaxConnLifetime)

	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		panic("Fail to parse pg env variables")
	}

	db.SetMaxOpenConns(maxConn)
	db.SetMaxIdleConns(maxIdle)
	db.SetConnMaxLifetime(time.Duration(lifetime) * time.Second)
	db.SetConnMaxIdleTime(time.Duration(idleLifetime) * time.Second)

	return &PG{db: db}
}

func (pg *PG) Close() {
	pg.db.Close()
}
