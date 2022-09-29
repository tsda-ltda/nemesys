package user

import (
	"context"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/auth"
	"github.com/fernandotsda/nemesys/api-manager/internal/roles"
	"github.com/fernandotsda/nemesys/shared/env"
)

func CreateDefaultUser(ctx context.Context, api *api.API) error {
	// check if user already exists
	sql := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`
	var e bool
	err := api.PgConn.QueryRow(ctx, sql, env.Username).Scan(&e)
	if err != nil {
		return err
	}
	if e {
		return nil
	}

	// hash password
	pwHashed, err := auth.Hash(env.PW, api.UserPWBcryptCost)
	if err != nil {
		return err
	}

	// save user in database
	sql = `INSERT INTO users (name, username, password, email, role) VALUES($1, $2, $3, $4, $5)`
	_, err = api.PgConn.Exec(ctx, sql,
		"Master",
		env.Username,
		pwHashed,
		"master@email.com",
		roles.Master,
	)
	if err != nil {
		return err
	}

	return nil
}
