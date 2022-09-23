package db

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type UsersORM struct {
	conn *pgx.Conn
}

// Creates the users table
func (u *UsersORM) CreateTable(ctx context.Context) (pgconn.CommandTag, error) {
	sql := `CREATE TABLE users (
		id serial4 PRIMARY KEY,
		name VARCHAR (50) NOT NULL,
		username VARCHAR (50) UNIQUE NOT NULL,
		password VARCHAR (50) NOT NULL,
		email VARCHAR (255) UNIQUE NOT NULL,
		role INT2 NOT NULL,
		teams_ids INT[]
	)`
	return u.conn.Exec(ctx, sql)
}

// Creates a new user
func (u *UsersORM) Create(ctx context.Context, user models.User) (pgconn.CommandTag, error) {
	sql := `INSERT INTO users (name, username, password, email, role, teams_ids)
		VALUES($1, $2, $3, $4, $5, $6)
	`
	return u.conn.Exec(ctx, sql,
		user.Name,
		user.Username,
		user.Password,
		user.Email,
		user.Role,
		user.TeamsIds,
	)
}

// Updates a existing user
func (u *UsersORM) Update(ctx context.Context, user models.User) (pgconn.CommandTag, error) {
	sql := `UPDATE users SET 
		(name, username, password, email, role, teams_ids) =
		($1, $2, $3, $4, $5, $6) WHERE id = $7
	`
	return u.conn.Exec(ctx, sql,
		user.Name,
		user.Username,
		user.Password,
		user.Email,
		user.Role,
		user.TeamsIds,
		user.Id,
	)
}

// Deletes a user
func (u *UsersORM) Delete(ctx context.Context, id int) (pgconn.CommandTag, error) {
	sql := `DELETE FROM users WHERE id = $1`
	return u.conn.Exec(ctx, sql, id)
}

// Check by id if user exists
func (u *UsersORM) Exists(ctx context.Context, id int) (bool, error) {
	var e bool
	sql := `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`
	err := u.conn.QueryRow(ctx, sql, id).Scan(&e)
	return e, err
}

// Check by username if user exists
func (u *UsersORM) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var e bool
	sql := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`
	err := u.conn.QueryRow(ctx, sql, username).Scan(&e)
	return e, err
}

// Check by email if user exists
func (u *UsersORM) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var e bool
	sql := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	err := u.conn.QueryRow(ctx, sql, email).Scan(&e)
	return e, err
}
