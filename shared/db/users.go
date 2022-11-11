package db

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/jackc/pgx/v5"
)

type Users struct {
	*pgx.Conn
}

// UsersExistsUsernameEmailResponse is the response for ExistsUsernameEmailResponse handler.
type UsersExistsUsernameEmailResponse struct {
	// UsernameExists is the username existence.
	UsernameExists bool
	// EmailExists is the email existence.
	EmailExists bool
}

// UsersLoginInfoResponse is the response for GetLoginInfo handler.
type UsersLoginInfoResponse struct {
	// Exists is the user existence.
	Exists bool
	// Id is the user id.
	Id int
	// Role is the user role.
	Role int
	// Password is the user password.
	Password string
}

// UsersGetWithoutPWResponse is the response for GetWithoutPW handler.
type UsersGetWithoutPWResponse struct {
	// Exists is the user existence.
	Exists bool
	// User is the user.
	User models.UserWithoutPW
}

// UsersGetRoleResponse is the response for GetRole handler.
type UsersGetRoleResponse struct {
	// Exists is the user existence.
	Exists bool
	// Role is the user role
	Role int16
}

const (
	sqlUsersExistsUsernameEmail = `SELECT 
		EXISTS (SELECT 1 FROM users WHERE username = $2 AND id != $1), 
		EXISTS (SELECT 1 FROM users WHERE email = $3 AND id != $1);`
	sqlUsersExistsUsername = `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1);`
	sqlUsersExists         = `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1);`
	sqlUsersCreate         = `INSERT INTO users (role, name, username, password, email) VALUES($1, $2, $3, $4, $5) RETURNING id;`
	sqlUsersUpdate         = `UPDATE users SET (role, name, username, password, email) = ($1, $2, $3, $4, $5) WHERE id = $6`
	sqlUsersDelete         = `DELETE FROM users WHERE id=$1;`
	sqlUsersMGetSimplified = `SELECT id, username, name FROM users LIMIT $1 OFFSET $2;`
	sqlUsersGetWithoutPW   = `SELECT username, name, email, role FROM users WHERE id = $1;`
	sqlUsersLoginInfo      = `SELECT id, role, password FROM users WHERE username = $1;`
	sqlUsersGetRole        = `SELECT role FROM users WHERE id = $1;`
	sqlUsersTeams          = `SELECT id, name, ident, descr FROM teams t 
		LEFT JOIN users_teams ut ON ut.team_id = t.id 
		WHERE ut.user_id = $1 LIMIT $2 OFFSET $3;`
)

// Exists return the existence of user.
func (c *Users) Exists(ctx context.Context, id int32) (exists bool, err error) {
	return exists, c.QueryRow(ctx, sqlUsersExists, id).Scan(&exists)
}

// ExistsUsername returns the existence of a user's username.
func (c *Users) ExistsUsername(ctx context.Context, username string) (exists bool, err error) {
	return exists, c.QueryRow(ctx, sqlUsersExistsUsername, username).Scan(&exists)
}

// ExistsUsernameEmail returns the existence of a user's username and email.
func (c *Users) ExistsUsernameEmail(ctx context.Context, username string, email string, userId int32) (r UsersExistsUsernameEmailResponse, err error) {
	return r, c.QueryRow(ctx, sqlUsersExistsUsernameEmail, userId, username, email).Scan(
		&r.UsernameExists,
		&r.EmailExists,
	)
}

// Create saves an user in database.
func (c *Users) Create(ctx context.Context, user models.User) (id int32, err error) {
	return id, c.QueryRow(ctx, sqlUsersCreate, user.Role, user.Name, user.Username, user.Password, user.Email).Scan(&id)
}

// Delete deletes a user by id if exists.
func (c *Users) Delete(ctx context.Context, id int32) (e bool, err error) {
	t, err := c.Exec(ctx, sqlUsersDelete, id)
	return t.RowsAffected() != 0, err
}

// MGetSimplified returns simplified users with a limit and a offset.
func (c *Users) MGetSimplified(ctx context.Context, limit int, offset int) (users []models.UserSimplified, err error) {
	users = []models.UserSimplified{}
	rows, err := c.Query(ctx, sqlUsersMGetSimplified, limit, offset)
	if err != nil {
		return users, err
	}
	defer rows.Close()
	for rows.Next() {
		var u models.UserSimplified
		err = rows.Scan(&u.Id, &u.Username, &u.Name)
		if err != nil {
			return users, err
		}
		users = append(users, u)
	}
	return users, nil
}

// GetWithout returns a user without password and it's existence.
func (c *Users) GetWithoutPW(ctx context.Context, id int32) (r UsersGetWithoutPWResponse, err error) {
	rows, err := c.Query(ctx, sqlUsersGetWithoutPW, id)
	if err != nil {
		return r, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(
			&r.User.Username,
			&r.User.Name,
			&r.User.Email,
			&r.User.Role,
		)
		if err != nil {
			return r, err
		}
		r.User.Id = id
		r.Exists = true
	}
	return r, nil
}

// Update updates a user if exists.
func (c *Users) Update(ctx context.Context, user models.User) (exists bool, err error) {
	t, err := c.Exec(ctx, sqlUsersUpdate,
		user.Role,
		user.Name,
		user.Username,
		user.Password,
		user.Email,
		user.Id,
	)
	if err != nil {
		return false, err
	}
	return t.RowsAffected() != 0, nil
}

// LoginInfo returns the information necessary to check a login attempt.
func (c *Users) LoginInfo(ctx context.Context, username string) (r UsersLoginInfoResponse, err error) {
	rows, err := c.Query(ctx, sqlUsersLoginInfo, username)
	if err != nil {
		return r, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(
			&r.Id,
			&r.Role,
			&r.Password,
		)
		if err != nil {
			return r, err
		}
		r.Exists = true
	}
	return r, nil
}

// GetRole returns an user's role if the user exists.
func (c *Users) GetRole(ctx context.Context, id int32) (r UsersGetRoleResponse, err error) {
	rows, err := c.Query(ctx, sqlUsersGetRole, id)
	if err != nil {
		return r, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&r.Role)
		if err != nil {
			return r, err
		}
		r.Exists = true
	}
	return r, nil
}

// Teams returns all user's teams with a limit and offset.
func (c *Users) Teams(ctx context.Context, userId int32, limit int, offset int) (teams []models.Team, err error) {
	teams = []models.Team{}
	rows, err := c.Query(ctx, sqlUsersTeams, userId, limit, offset)
	if err != nil {
		return teams, err
	}
	defer rows.Close()
	for rows.Next() {
		var t models.Team
		err = rows.Scan(&t.Id, &t.Name, &t.Ident, &t.Descr)
		if err != nil {
			return teams, err
		}
		teams = append(teams, t)
	}
	return teams, nil
}
