package db

var databaseEmpty = false

func (pg *PgConn) Init() error {
	if !databaseEmpty {
		return nil
	}

	// create users table
	_, err := pg.CreateUserTable()
	if err != nil {
		return err
	}

	return nil
}
