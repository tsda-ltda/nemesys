package db

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/jackc/pgx/v5"
)

type Users struct {
	*pgx.Conn
}

type LoginInfo struct {
	Id     int
	Role   int
	PW     string
	Exists bool
}

const (
	sqlUsersExists              = `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1);`
	sqlUsersExistsUsername      = `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1);`
	sqlUsersExistsUsernameEmail = `SELECT EXISTS (
		SELECT 1 FROM users WHERE username = $1
		) as EX1, EXISTS (
			SELECT 1 FROM users WHERE email = $2
		) as EX2;`
	sqlUsersCreate         = `INSERT INTO users (role, name, username, password, email) VALUES($1, $2, $3, $4, $5);`
	sqlUsersUpdate         = `UPDATE users SET (role, name, username, password, email) = ($1, $2, $3, $4, $5) WHERE id = $6`
	sqlUsersDelete         = `DELETE FROM users WHERE id=$1;`
	sqlUsersMGetSimplified = `SELECT id, username, name FROM users LIMIT $1 OFFSET $2;`
	sqlUsersGetWithoutPW   = `SELECT username, name, email, role FROM users WHERE id = $1;`
	sqlUsersLoginInfo      = `SELECT id, role, password FROM users WHERE username = $1;`
	sqlUsersGetRole        = `SELECT role FROM users WHERE id = $1;`

	sqlUsersUsernameEmailAvailableUpdate = `SELECT 
		EXISTS (SELECT 1 FROM users WHERE  id != $1 AND username = $2),
		EXISTS (SELECT 1 FROM users WHERE id != $1 AND username = $3);`
	sqlUsersTeams = `SELECT id, name, ident, descr FROM teams t 
		LEFT JOIN users_teams ut ON ut.team_id = t.id 
		WHERE ut.user_id = $1 LIMIT $2 OFFSET $3;`
)

// Exists return the existence of user. Returns an error if fails to check.
func (c *Users) Exists(ctx context.Context, id int) (e bool, err error) {
	err = c.QueryRow(ctx, sqlUsersExists, id).Scan(&e)
	return e, err
}

// ExistsUsername returns the existence of a user's username.
// Returns an error if fails to check.
func (c *Users) ExistsUsername(ctx context.Context, username string) (e bool, err error) {
	err = c.QueryRow(ctx, sqlUsersExistsUsername, username).Scan(&e)
	return e, err
}

// ExistsUsernameEmail returns the existence of a user's username and email.
// Returns an error if fails to check.
func (c *Users) ExistsUsernameEmail(ctx context.Context, username string, email string) (ue bool, ee bool, err error) {
	err = c.QueryRow(ctx, sqlUsersExistsUsernameEmail, username, email).Scan(&ue, &ee)
	return ue, ee, err
}

// Create saves an user in database. Returns an err if fails to create.
func (c *Users) Create(ctx context.Context, user models.User) error {
	_, err := c.Exec(ctx, sqlUsersCreate, user.Role, user.Name, user.Username, user.Password, user.Email)
	return err
}

// Delete deletes a user by id if exists. Returns an error if fails to delete.
func (c *Users) Delete(ctx context.Context, id int) (e bool, err error) {
	t, err := c.Exec(ctx, sqlUsersDelete, id)
	return t.RowsAffected() != 0, err
}

// MGetSimplified returns simplified users with a limit and a offset.
// Returns an error if fails to get users.
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
// Returns an error if fail to get user.
func (c *Users) GetWithoutPW(ctx context.Context, id int) (user models.UserWithoutPW, e bool, err error) {
	rows, err := c.Query(ctx, sqlUsersGetWithoutPW, id)
	if err != nil {
		return user, false, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(
			&user.Username,
			&user.Name,
			&user.Email,
			&user.Role,
		)
		if err != nil {
			return user, false, err
		}
		user.Id = id
		e = true
	}
	return user, e, nil
}

// UsernameEmailAvailableToUpdate returns if a username and a email is available to update.
// Returns an error if fail to check.
func (c *Users) UsernameEmailAvailableToUpdate(ctx context.Context, id int, username string, email string) (ue bool, ee bool, err error) {
	err = c.QueryRow(ctx, sqlUsersUsernameEmailAvailableUpdate, id, username, email).Scan(&ue, &ee)
	return ue, ee, err
}

// Update updates a user if exists. Returns an error if fail to update user.
func (c *Users) Update(ctx context.Context, user models.User) (e bool, err error) {
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
// Returns an error if fails to get the information.
func (c *Users) LoginInfo(ctx context.Context, username string) (li LoginInfo, err error) {
	rows, err := c.Query(ctx, sqlUsersLoginInfo, username)
	if err != nil {
		return li, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&li.Id, &li.Role, &li.PW)
		if err != nil {
			return li, err
		}
		li.Exists = true
	}
	return li, nil
}

// GetRole returns an user's role if the user exists.
// Returns an error if fails to get role.
func (c *Users) GetRole(ctx context.Context, id int) (e bool, role int, err error) {
	rows, err := c.Query(ctx, sqlUsersGetRole, id)
	if err != nil {
		return false, role, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&role)
		if err != nil {
			return false, role, err
		}
		e = true
	}
	return e, role, nil
}

// Teams returns all user's teams with a limit and offset.
// Returns an error if fail to get teams.
func (c *Users) Teams(ctx context.Context, userId int, limit int, offset int) (teams []models.Team, err error) {
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
