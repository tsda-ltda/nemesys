package db

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/jackc/pgx/v5/pgconn"
)

func (pg *PgConn) CreateUserTable() (pgconn.CommandTag, error) {
	sql := `CREATE TABLE users (
		id serial4 PRIMARY KEY,
		name VARCHAR (50) NOT NULL,
		username VARCHAR (50) UNIQUE NOT NULL,
		password VARCHAR (50) NOT NULL,
		email VARCHAR (255) UNIQUE NOT NULL,
		role INT2 NOT NULL,
		teams_ids INT[]
	)`
	return pg.Exec(context.Background(), sql)
}

func (pg *PgConn) CreateUser(user models.User) (pgconn.CommandTag, error) {
	sql := `INSERT INTO users (name, username, password, email, role, teams_ids)
		VALUES($1, $2, $3, $4, $5, $6)
	`
	return pg.Exec(context.Background(), sql,
		user.Name,
		user.Username,
		user.Password,
		user.Email,
		user.Role,
		user.TeamsIds,
	)
}
