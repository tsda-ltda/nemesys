package pg

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/models"
)

var TeamValidOrderByColumns = []string{"ident", "descr", "name"}

type TeamQueryFilters struct {
	Ident     string `type:"ilike" column:"ident"`
	Name      string `type:"ilike" column:"name"`
	Descr     string `type:"ilike" column:"descr"`
	OrderBy   string
	OrderByFn string
}

func (f TeamQueryFilters) GetOrderBy() string {
	return f.OrderBy
}

func (f TeamQueryFilters) GetOrderByFn() string {
	return f.OrderByFn
}

// TeamsExistsRelUserTeamResponse is the response for ExistsRelUserTeam handler.
type TeamsExistsRelUserTeamResponse struct {
	// RelationExist is the relation existence.
	RelationExist bool
	// UserExists is the user existence.
	UserExists bool
	// TeamExists is the team existence.
	TeamExists bool
}

const (
	sqlTeamsCreate      = `INSERT INTO teams (ident, descr, name) VALUES($1, $2, $3) RETURNING id;`
	sqlTeamsExistsIdent = `SELECT EXISTS (SELECT 1 FROM teams WHERE ident = $1 AND id != $2);`
	sqlTeamsDelete      = `DELETE FROM teams WHERE id = $1;`
	sqlTeamsGet         = `SELECT ident, descr, name FROM teams WHERE id = $1;`
	sqlTeamsGetByIdent  = `SELECT id, descr, name FROM teams WHERE ident = $1;`
	sqlTeamsUpdate      = `UPDATE teams SET (name, ident, descr) = ($1, $2, $3) WHERE id = $4;`
	sqlTeamsAddMember   = `INSERT INTO users_teams (user_id, team_id) VALUES ($1, $2);`
	sqlTeamsRemMember   = `DELETE FROM users_teams WHERE user_id = $1 AND team_id = $2;`
	sqlTeamsMGetMembers = `SELECT u.id, u.first_name, u.last_name, u.username, u.role, u.email 
		FROM users u 
		LEFT JOIN users_teams ut ON ut.user_id = u.id WHERE ut.team_id = $1
		LIMIT $2 OFFSET $3;`
	sqlTeamsExistsRelUserTeam = `SELECT 
		EXISTS(SELECT 1 FROM users_teams WHERE user_id = $1 AND team_id = $2), 
		EXISTS(SELECT 1 FROM users WHERE id=$1), 
		EXISTS(SELECT 1 FROM teams WHERE id=$2);`
	sqlTeamsMemberExists = `SELECT EXISTS (SELECT 1 FROM users_teams WHERE user_id = $1 AND team_id = $2);`

	customSqlTeamsMGet = `SELECT id, name, descr, ident FROM teams %s LIMIT $1 OFFSET $2`
)

func (pg *PG) TeamIdentExists(ctx context.Context, ident string, id int32) (exists bool, err error) {
	return exists, pg.db.QueryRowContext(ctx, sqlTeamsExistsIdent, ident, id).Scan(&exists)
}

func (pg *PG) CreateTeam(ctx context.Context, team models.Team) (id int32, err error) {
	return id, pg.db.QueryRowContext(ctx, sqlTeamsCreate,
		team.Ident,
		team.Descr,
		team.Name,
	).Scan(&id)
}

func (pg *PG) DeleteTeam(ctx context.Context, id int32) (exists bool, err error) {
	t, err := pg.db.ExecContext(ctx, sqlTeamsDelete, id)
	if err != nil {
		return false, err
	}
	rowsAffected, _ := t.RowsAffected()
	return rowsAffected != 0, err
}

func (pg *PG) GetTeam(ctx context.Context, id int32) (exists bool, team models.Team, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlTeamsGet, id)
	if err != nil {
		return exists, team, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(
			&team.Ident,
			&team.Descr,
			&team.Name,
		)
		if err != nil {
			return false, team, err
		}
		team.Id = id
		exists = true
	}
	return exists, team, nil
}

func (pg *PG) GetTeams(ctx context.Context, filters TeamQueryFilters, limit int, offset int) (teams []models.Team, err error) {
	sql, err := applyFilters(filters, customSqlTeamsMGet, TeamValidOrderByColumns)
	if err != nil {
		return nil, err
	}
	rows, err := pg.db.QueryContext(ctx, sql, limit, offset)
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
			&t.Descr,
			&t.Ident,
		)
		if err != nil {
			return nil, err
		}
		teams = append(teams, t)
	}
	return teams, nil
}

func (pg *PG) UpdateTeam(ctx context.Context, team models.Team) (exists bool, err error) {
	t, err := pg.db.ExecContext(ctx, sqlTeamsUpdate,
		team.Name,
		team.Ident,
		team.Descr,
		team.Id,
	)
	if err != nil {
		return false, err
	}
	rowsAffected, _ := t.RowsAffected()
	return rowsAffected != 0, err
}

func (pg *PG) ExistsRelUserTeam(ctx context.Context, userId int32, teamId int32) (r TeamsExistsRelUserTeamResponse, err error) {
	return r, pg.db.QueryRowContext(ctx, sqlTeamsExistsRelUserTeam, userId, teamId).Scan(
		&r.RelationExist,
		&r.UserExists,
		&r.TeamExists,
	)
}

func (pg *PG) AddTeamMember(ctx context.Context, userId int32, teamId int32) error {
	_, err := pg.db.ExecContext(ctx, sqlTeamsAddMember, userId, teamId)
	return err
}

func (pg *PG) RemoveTeamMember(ctx context.Context, userId int32, teamId int32) (exists bool, err error) {
	t, err := pg.db.ExecContext(ctx, sqlTeamsRemMember, userId, teamId)
	if err != nil {
		return false, err
	}
	rowsAffected, _ := t.RowsAffected()
	return rowsAffected != 0, err
}

func (pg *PG) GetTeamMembers(ctx context.Context, teamId int32, limit int, offset int) (users []models.User, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlTeamsMGetMembers, teamId, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	users = make([]models.User, 0, limit)
	var u models.User
	for rows.Next() {
		err = rows.Scan(
			&u.Id,
			&u.FirstName,
			&u.LastName,
			&u.Username,
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

func (pg *PG) TeamMemberExists(ctx context.Context, teamId int32, userId int32) (exists bool, err error) {
	return exists, pg.db.QueryRowContext(ctx, sqlTeamsMemberExists, userId, teamId).Scan(&exists)
}
