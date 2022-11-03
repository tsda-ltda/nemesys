package db

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/jackc/pgx/v5"
)

type Teams struct {
	*pgx.Conn
}

const (
	sqlTeamsCreate               = `INSERT INTO teams (ident, descr, name) VALUES($1, $2, $3);`
	sqlTeamsExistsIdent          = `SELECT EXISTS (SELECT 1 FROM teams WHERE ident = $1);`
	sqlTeamsDelete               = `DELETE FROM teams WHERE id = $1;`
	sqlTeamsGet                  = `SELECT ident, descr, name FROM teams WHERE id = $1;`
	sqlTeamsGetByIdent           = `SELECT id, descr, name FROM teams WHERE ident = $1;`
	sqlTeamsMGet                 = `SELECT id, name, descr, ident FROM teams LIMIT $1 OFFSET $2;`
	sqlTeamsIdentAvailableUpdate = `SELECT EXISTS (SELECT 1 FROM teams WHERE id != $1 AND ident = $2);`
	sqlTeamsUpdate               = `UPDATE teams SET (name, ident, descr) = ($1, $2, $3) WHERE id = $4;`
	sqlTeamsAddMember            = `INSERT INTO users_teams (user_id, team_id) VALUES ($1, $2);`
	sqlTeamsRemMember            = `DELETE FROM users_teams WHERE user_id = $1 AND team_id = $2;`
	sqlTeamsMGetMembers          = `SELECT u.id, u.name, u.username 
		FROM users u 
		LEFT JOIN users_teams ut ON ut.user_id = u.id WHERE ut.team_id = $1
		LIMIT $2 OFFSET $3;`
	sqlTeamsExistsRelUserTeam = `SELECT 
		EXISTS(SELECT 1 FROM users_teams WHERE user_id = $1 AND team_id = $2), 
		EXISTS(SELECT 1 FROM users WHERE id=$1), 
		EXISTS(SELECT 1 FROM teams WHERE id=$2);`
)

// ExistsIdent returns the existence of a team's ident.
// Returns an error if fails to check.
func (c *Teams) ExistsIdent(ctx context.Context, ident string) (e bool, err error) {
	err = c.QueryRow(ctx, sqlTeamsExistsIdent, ident).Scan(&e)
	return e, err
}

// Create creates a team. Returns an error if fails to create.
func (c *Teams) Create(ctx context.Context, team models.Team) error {
	_, err := c.Exec(ctx, sqlTeamsCreate,
		team.Ident,
		team.Descr,
		team.Name,
	)
	return err
}

// Delete deletes a team if exists. Returns an error if fails to delete.
func (c *Teams) Delete(ctx context.Context, id int32) (e bool, err error) {
	t, err := c.Exec(ctx, sqlTeamsDelete, id)
	return t.RowsAffected() != 0, err
}

// Get returns a team by id if exists. Returns an error if fails to get.
func (c *Teams) Get(ctx context.Context, id int32) (e bool, team models.Team, err error) {
	rows, err := c.Query(ctx, sqlTeamsGet, id)
	if err != nil {
		return false, team, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&team.Ident, &team.Descr, &team.Name)
		if err != nil {
			return false, team, err
		}
		team.Id = id
		e = true
	}
	return e, team, nil
}

// MGet returns an array of teams with a limit and offset.
// Returns an error if fails to get teams.
func (c *Teams) MGet(ctx context.Context, limit int, offset int) (teams []models.Team, err error) {
	teams = []models.Team{}
	rows, err := c.Query(ctx, sqlTeamsMGet, limit, offset)
	if err != nil {
		return teams, err
	}
	defer rows.Close()
	for rows.Next() {
		var t models.Team
		err = rows.Scan(
			&t.Id,
			&t.Name,
			&t.Descr,
			&t.Ident,
		)
		if err != nil {
			return teams, err
		}
		teams = append(teams, t)
	}
	return teams, nil
}

// IdentAvailableUpdate returns the availability of a team's ident.
// Returns an error if fails to check availability.
func (c *Teams) IdentAvailableUpdate(ctx context.Context, id int32, ident string) (e bool, err error) {
	err = c.QueryRow(ctx, sqlTeamsIdentAvailableUpdate, id, ident).Scan(&e)
	return e, err
}

// Update updates a team if exists. Returns an error if fails to update.
func (c *Teams) Update(ctx context.Context, team models.Team) (e bool, err error) {
	t, err := c.Exec(ctx, sqlTeamsUpdate,
		team.Name,
		team.Ident,
		team.Descr,
		team.Id,
	)
	return t.RowsAffected() != 0, err
}

// ExistsRelUserTeam returns a user-team relation, user and team existence.
// Returns an error if fails to check.
func (c *Teams) ExistsRelUserTeam(ctx context.Context, userId int32, teamId int) (rel bool, ue bool, te bool, err error) {
	err = c.QueryRow(ctx, sqlTeamsExistsRelUserTeam, userId, teamId).Scan(&rel, &ue, &te)
	return rel, ue, te, err
}

// AddMember add a user to a team. Returns an error if fails to add.
func (c *Teams) AddMember(ctx context.Context, userId int32, teamId int32) error {
	_, err := c.Exec(ctx, sqlTeamsAddMember, userId, teamId)
	return err
}

// RemMember removes a memeber from a team if relation exists.
// Returns an error if fails to remove.
func (c *Teams) RemMember(ctx context.Context, userId int32, teamId int32) (e bool, err error) {
	t, err := c.Exec(ctx, sqlTeamsRemMember, userId, teamId)
	return t.RowsAffected() != 0, err
}

// MGetMembers returns all members of a team with a limit and offset.
// Returns an error if fails to get.
func (c *Teams) MGetMembers(ctx context.Context, teamId int32, limit int, offset int) (m []models.UserSimplified, err error) {
	m = []models.UserSimplified{}
	rows, err := c.Query(ctx, sqlTeamsMGetMembers, teamId, limit, offset)
	if err != nil {
		return m, err
	}
	defer rows.Close()
	for rows.Next() {
		var u models.UserSimplified
		err = rows.Scan(&u.Id, &u.Name, &u.Username)
		if err != nil {
			return m, err
		}
		m = append(m, u)
	}
	return m, nil
}
