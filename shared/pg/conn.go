package pg

import (
	"database/sql"
	"fmt"
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

	// maxConns, err := strconv.Atoi(env.PGMaxConns)
	// if err != nil {
	// 	panic("fail to parse maxConns")
	// }
	// maxIdleConns, err := strconv.Atoi(env.PGMaxIdleConns)
	// if err != nil {
	// 	panic("fail to parse maxConns")
	// }
	// maxConnLifetime, err := strconv.Atoi(env.PGConnMaxLifetime)
	// if err != nil {
	// 	panic("fail to parse maxConnLifetime")
	// }

	// db.SetConnMaxLifetime(time.Duration(maxConnLifetime) * time.Minute)
	// db.SetMaxIdleConns(maxIdleConns)
	// db.SetMaxOpenConns(maxConns)
	// Maximum Idle Connections
	db.SetMaxIdleConns(5)
	// Maximum Open Connections
	db.SetMaxOpenConns(10)
	// Idle Connection Timeout
	db.SetConnMaxIdleTime(1 * time.Second)
	// Connection Lifetime
	db.SetConnMaxLifetime(30 * time.Second)
	return &PG{db: db}
}

func (pg *PG) Close() {
	pg.db.Close()
}
