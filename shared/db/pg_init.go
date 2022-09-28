package db

import "context"

func (pg *PgConn) Init() error {
	ctx := context.Background()

	// create users table
	_, err := pg.Users.CreateTable(ctx)
	if err != nil {
		return err
	}

	// create teams table
	_, err = pg.Teams.CreateTable(ctx)
	if err != nil {
		return err
	}

	return nil
}
