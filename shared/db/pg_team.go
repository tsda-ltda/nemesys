package db

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type TeamsORM struct {
	conn *pgx.Conn
}

// Creates the teams table
func (u *TeamsORM) CreateTable(ctx context.Context) (pgconn.CommandTag, error) {
	sql := `CREATE TABLE teams (
		id serial4 PRIMARY KEY,
		name VARCHAR (50) NOT NULL,
		ident VARCHAR (50) UNIQUE NOT NULL,
		users_ids INT[] NOT NULL,
	)`
	return u.conn.Exec(ctx, sql)
}

// Creates a new team
func (u *TeamsORM) Create(ctx context.Context, team models.Team) (pgconn.CommandTag, error) {
	sql := `INSERT INTO teams (name, ident, users_ids)
		VALUES($1, $2, $3)
	`
	return u.conn.Exec(ctx, sql,
		team.Name,
		team.Ident,
		team.UsersIds,
	)
}

// Updates a team
func (u *TeamsORM) Update(ctx context.Context, team models.Team) (pgconn.CommandTag, error) {
	sql := `UPDATE teams SET
		(name, users_ids, ident) =
		($1, $2, $3) WHERE id = $4
	`
	return u.conn.Exec(ctx, sql,
		team.Name,
		team.UsersIds,
		team.Ident,
		team.Id,
	)
}

// Deletes a team
func (u *TeamsORM) Delete(ctx context.Context, id int) (pgconn.CommandTag, error) {
	sql := `DELETE FROM teams WHERE id = $1`
	return u.conn.Exec(ctx, sql, id)
}

// Check by id if team exists
func (u *TeamsORM) Exists(ctx context.Context, id int) (bool, error) {
	var e bool
	sql := `SELECT EXISTS(SELECT 1 FROM teams WHERE id = $1)`
	err := u.conn.QueryRow(ctx, sql, id).Scan(&e)
	return e, err
}
