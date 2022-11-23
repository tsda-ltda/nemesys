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
	exists, err := api.PG.UsernameExists(ctx, env.DefaultUsername)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	pwHashed, err := auth.Hash(env.DefaultPassword, api.UserPWBcryptCost)
	if err != nil {
		return err
	}

	_, err = api.PG.CreateUser(ctx, models.User{
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
