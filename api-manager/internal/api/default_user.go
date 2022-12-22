package api

import (
	"context"

	"github.com/fernandotsda/nemesys/api-manager/internal/auth"
	"github.com/fernandotsda/nemesys/api-manager/internal/roles"
	"github.com/fernandotsda/nemesys/shared/env"
	"github.com/fernandotsda/nemesys/shared/models"
)

func (api *API) createDefaultUser(ctx context.Context) error {
	usersLen, err := api.PG.CountUsersWithLimit(ctx, 1)
	if err != nil {
		return err
	}
	if usersLen > 0 {
		return nil
	}

	pwHashed, err := auth.Hash(env.DefaultPassword, api.UserPWBcryptCost)
	if err != nil {
		return err
	}

	_, err = api.PG.CreateUser(ctx, models.User{
		Role:      roles.Master,
		FirstName: "Default",
		LastName:  "Master",
		Username:  env.DefaultUsername,
		Password:  pwHashed,
		Email:     "master@master.com",
	})
	if err != nil {
		return err
	}
	api.Log.Info("Default master user created")

	return nil
}
