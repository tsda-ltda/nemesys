package db

import "context"

var databaseEmpty = false

func (pg *PgConn) Init() error {
	if !databaseEmpty {
		return nil
	}
	ctx := context.Background()

	// create users table
	_, err := pg.Users.CreateTable(ctx)
	if err != nil {
		return err
	}

	return nil
}
