package pg

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/models"
)

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

const (
	sqlUsersCountWithLimit      = `SELECT COUNT(*) FROM users LIMIT $1;`
	sqlUsersExistsUsernameEmail = `SELECT 
		EXISTS (SELECT 1 FROM users WHERE username = $2 AND id != $1), 
		EXISTS (SELECT 1 FROM users WHERE email = $3 AND id != $1);`
	sqlUsersExistsUsername = `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1);`
	sqlUsersExists         = `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1);`
	sqlUsersCreate         = `INSERT INTO users (role, first_name, last_name, username, password, email) VALUES($1, $2, $3, $4, $5, $6) RETURNING id;`
	sqlUsersUpdate         = `UPDATE users SET (role, first_name, last_name, username, password, email) = ($1, $2, $3, $4, $5, $6) WHERE id = $6`
	sqlUsersDelete         = `DELETE FROM users WHERE id=$1;`
	sqlUsersMGet           = `SELECT id, username, first_name, last_name, role, email FROM users LIMIT $1 OFFSET $2;`
	sqlUsersGetWithoutPW   = `SELECT username, first_name, last_name, email, role FROM users WHERE id = $1;`
	sqlUsersLoginInfo      = `SELECT id, role, password FROM users WHERE username = $1;`
	sqlUsersGetRole        = `SELECT role FROM users WHERE id = $1;`
	sqlUsersTeams          = `SELECT id, name, ident, descr FROM teams t 
		LEFT JOIN users_teams ut ON ut.team_id = t.id 
		WHERE ut.user_id = $1 LIMIT $2 OFFSET $3;`
)

func (pg *PG) CountUsersWithLimit(ctx context.Context, limit int) (users int, err error) {
	return users, pg.db.QueryRowContext(ctx, sqlUsersCountWithLimit, limit).Scan(&users)
}

func (pg *PG) UserExists(ctx context.Context, id int32) (exists bool, err error) {
	return exists, pg.db.QueryRowContext(ctx, sqlUsersExists, id).Scan(&exists)
}

func (pg *PG) UsernameExists(ctx context.Context, username string) (exists bool, err error) {
	return exists, pg.db.QueryRowContext(ctx, sqlUsersExistsUsername, username).Scan(&exists)
}

func (pg *PG) UsernameAndEmailExists(ctx context.Context, username string, email string, userId int32) (usernameExists bool, emailExists bool, err error) {
	return usernameExists, emailExists, pg.db.QueryRowContext(ctx, sqlUsersExistsUsernameEmail, userId, username, email).Scan(
		&usernameExists,
		&emailExists,
	)
}

func (pg *PG) CreateUser(ctx context.Context, user models.User) (id int32, err error) {
	return id, pg.db.QueryRowContext(ctx, sqlUsersCreate, user.Role, user.FirstName, user.LastName, user.Username, user.Password, user.Email).Scan(&id)
}

func (pg *PG) DeleteUser(ctx context.Context, id int32) (e bool, err error) {
	t, err := pg.db.ExecContext(ctx, sqlUsersDelete, id)
	if err != nil {
		return false, err
	}
	rowsAffected, _ := t.RowsAffected()
	return rowsAffected != 0, err
}

func (pg *PG) GetUsers(ctx context.Context, limit int, offset int) (users []models.User, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlUsersMGet, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	users = make([]models.User, 0, limit)
	var u models.User
	for rows.Next() {
		err = rows.Scan(
			&u.Id,
			&u.Username,
			&u.FirstName,
			&u.LastName,
			&u.Role,
			&u.Email,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (pg *PG) GetUserWithoutPW(ctx context.Context, id int32) (exists bool, user models.UserWithoutPW, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlUsersGetWithoutPW, id)
	if err != nil {
		return false, user, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(
			&user.Username,
			&user.FirstName,
			&user.LastName,
			&user.Email,
			&user.Role,
		)
		if err != nil {
			return false, user, err
		}
		user.Id = id
		exists = true
	}
	return exists, user, nil
}

func (pg *PG) UpdateUser(ctx context.Context, user models.User) (exists bool, err error) {
	t, err := pg.db.ExecContext(ctx, sqlUsersUpdate,
		user.Role,
		user.FirstName,
		user.LastName,
		user.Username,
		user.Password,
		user.Email,
		user.Id,
	)
	if err != nil {
		return false, err
	}
	rowsAffected, _ := t.RowsAffected()
	return rowsAffected != 0, err
}

func (pg *PG) GetLoginInfo(ctx context.Context, username string) (r UsersLoginInfoResponse, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlUsersLoginInfo, username)
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

func (pg *PG) GetUserRole(ctx context.Context, id int32) (exists bool, role int16, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlUsersGetRole, id)
	if err != nil {
		return exists, role, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&role)
		if err != nil {
			return exists, role, err
		}
		exists = true
	}
	return exists, role, nil
}

func (pg *PG) GetUserTeams(ctx context.Context, userId int32, limit int, offset int) (teams []models.Team, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlUsersTeams, userId, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	teams = make([]models.Team, 0, limit)
	var t models.Team
	for rows.Next() {
		err = rows.Scan(
			&t.Id,
			&t.Name,
			&t.Ident,
			&t.Descr,
		)
		if err != nil {
			return nil, err
		}
		teams = append(teams, t)
	}
	return teams, nil
}
