package pg

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/models"
)

// TeamsGetResponse is the response for Get handler.
type TeamsGetResponse struct {
	// Exists is the team existence.
	Exists bool
	// Team is the team.
	Team models.Team
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
	sqlTeamsMGet        = `SELECT id, name, descr, ident FROM teams LIMIT $1 OFFSET $2;`
	sqlTeamsUpdate      = `UPDATE teams SET (name, ident, descr) = ($1, $2, $3) WHERE id = $4;`
	sqlTeamsAddMember   = `INSERT INTO users_teams (user_id, team_id) VALUES ($1, $2);`
	sqlTeamsRemMember   = `DELETE FROM users_teams WHERE user_id = $1 AND team_id = $2;`
	sqlTeamsMGetMembers = `SELECT u.id, u.name, u.username 
		FROM users u 
		LEFT JOIN users_teams ut ON ut.user_id = u.id WHERE ut.team_id = $1
		LIMIT $2 OFFSET $3;`
	sqlTeamsExistsRelUserTeam = `SELECT 
		EXISTS(SELECT 1 FROM users_teams WHERE user_id = $1 AND team_id = $2), 
		EXISTS(SELECT 1 FROM users WHERE id=$1), 
		EXISTS(SELECT 1 FROM teams WHERE id=$2);`
)

func (pg *PG) TeamIdentExists(ctx context.Context, ident string, id int32) (exists bool, err error) {
	c, err := pg.pool.Acquire(ctx)
	if err != nil {
		return false, err
	}
	defer c.Release()
	return exists, c.QueryRow(ctx, sqlTeamsExistsIdent, ident, id).Scan(&exists)
}

func (pg *PG) CreateTeam(ctx context.Context, team models.Team) (id int32, err error) {
	c, err := pg.pool.Acquire(ctx)
	if err != nil {
		return id, err
	}
	defer c.Release()
	return id, c.QueryRow(ctx, sqlTeamsCreate,
		team.Ident,
		team.Descr,
		team.Name,
	).Scan(&id)
}

func (pg *PG) DeleteTeam(ctx context.Context, id int32) (exists bool, err error) {
	c, err := pg.pool.Acquire(ctx)
	if err != nil {
		return false, err
	}
	defer c.Release()
	t, err := c.Exec(ctx, sqlTeamsDelete, id)
	return t.RowsAffected() != 0, err
}

func (pg *PG) GetTeam(ctx context.Context, id int32) (r TeamsGetResponse, err error) {
	c, err := pg.pool.Acquire(ctx)
	if err != nil {
		return r, err
	}
	defer c.Release()
	rows, err := c.Query(ctx, sqlTeamsGet, id)
	if err != nil {
		return r, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(
			&r.Team.Ident,
			&r.Team.Descr,
			&r.Team.Name,
		)
		if err != nil {
			return r, err
		}
		r.Team.Id = id
		r.Exists = true
	}
	return r, nil
}

func (pg *PG) GetTeams(ctx context.Context, limit int, offset int) (teams []models.Team, err error) {
	c, err := pg.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Release()
	teams = []models.Team{}
	rows, err := c.Query(ctx, sqlTeamsMGet, limit, offset)
	if err != nil {
		return nil, err
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
			return nil, err
		}
		teams = append(teams, t)
	}
	return teams, nil
}

func (pg *PG) UpdateTeam(ctx context.Context, team models.Team) (exists bool, err error) {
	c, err := pg.pool.Acquire(ctx)
	if err != nil {
		return false, err
	}
	defer c.Release()
	t, err := c.Exec(ctx, sqlTeamsUpdate,
		team.Name,
		team.Ident,
		team.Descr,
		team.Id,
	)
	return t.RowsAffected() != 0, err
}

func (pg *PG) ExistsRelUserTeam(ctx context.Context, userId int32, teamId int32) (r TeamsExistsRelUserTeamResponse, err error) {
	c, err := pg.pool.Acquire(ctx)
	if err != nil {
		return r, err
	}
	defer c.Release()
	return r, c.QueryRow(ctx, sqlTeamsExistsRelUserTeam, userId, teamId).Scan(
		&r.RelationExist,
		&r.UserExists,
		&r.TeamExists,
	)
}

func (pg *PG) AddTeamMember(ctx context.Context, userId int32, teamId int32) error {
	c, err := pg.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer c.Release()
	_, err = c.Exec(ctx, sqlTeamsAddMember, userId, teamId)
	return err
}

func (pg *PG) RemoveTeamMember(ctx context.Context, userId int32, teamId int32) (exists bool, err error) {
	c, err := pg.pool.Acquire(ctx)
	if err != nil {
		return false, err
	}
	defer c.Release()
	t, err := c.Exec(ctx, sqlTeamsRemMember, userId, teamId)
	return t.RowsAffected() != 0, err
}

func (pg *PG) GetTeamMembers(ctx context.Context, teamId int32, limit int, offset int) (users []models.UserSimplified, err error) {
	c, err := pg.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Release()
	users = []models.UserSimplified{}
	rows, err := c.Query(ctx, sqlTeamsMGetMembers, teamId, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var u models.UserSimplified
		err = rows.Scan(&u.Id, &u.Name, &u.Username)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}
