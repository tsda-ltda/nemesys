package user

import (
	"context"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/auth"
	"github.com/fernandotsda/nemesys/api-manager/internal/roles"
	"github.com/fernandotsda/nemesys/shared/env"
	"github.com/fernandotsda/nemesys/shared/models"
)

func CreateDefaultUser(ctx context.Context, api *api.API) error {
	// check if user already exists
	e, err := api.PgConn.Users.ExistsUsername(ctx, env.DefaultUsername)
	if err != nil {
		return err
	}
	if e {
		return nil
	}

	// hash password
	pwHashed, err := auth.Hash(env.DefaultPassword, api.UserPWBcryptCost)
	if err != nil {
		return err
	}

	// save user
	err = api.PgConn.Users.Create(ctx, models.User{
		Role:     roles.Master,
		Name:     "Default Master",
		Username: env.DefaultUsername,
		Password: pwHashed,
		Email:    "master@master.com",
	})
	if err != nil {
		return err
	}

	return nil
}
