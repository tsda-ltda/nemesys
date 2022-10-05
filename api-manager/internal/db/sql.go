package db

const (
	// Creates the users table if not exists.
	sqlCreateUsersTable = `CREATE TABLE IF NOT EXISTS users (
			id serial4 PRIMARY KEY,
			name VARCHAR (50) NOT NULL,
			username VARCHAR (50) UNIQUE NOT NULL,
			password VARCHAR (255) NOT NULL,
			email VARCHAR (255) UNIQUE NOT NULL,
			role INT2 NOT NULL
	);`

	// Creates the team table if not exists.
	sqlCreateTeamsTable = `CREATE TABLE IF NOT EXISTS teams (
			id serial4 PRIMARY KEY,
			name VARCHAR (50) NOT NULL,
			ident VARCHAR (50) UNIQUE NOT NULL,
			descr VARCHAR (255) NOT NULL
	);`

	// Creates the team-users relation table if not exists
	sqlCreateUsersTeamsTable = `CREATE TABLE IF NOT EXISTS users_teams (
			userId int,
			teamId int,

			CONSTRAINT fk_userId
				FOREIGN KEY(userId)
					REFERENCES users(id)
					ON DELETE CASCADE,
			CONSTRAINT fk_teamId
				FOREIGN KEY(teamId)
					REFERENCES teams(id)
					ON DELETE CASCADE
	);`
)
