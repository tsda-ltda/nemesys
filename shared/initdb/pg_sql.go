package initdb

const (
	// Creates the users table if not exists.
	sqlCreateUsersTable = `CREATE TABLE users (
			id serial4 PRIMARY KEY,
			name VARCHAR (50) NOT NULL,
			username VARCHAR (50) UNIQUE NOT NULL,
			password VARCHAR (255) NOT NULL,
			email VARCHAR (255) UNIQUE NOT NULL,
			role INT2 NOT NULL
	);`

	// Creates the team table if not exists.
	sqlCreateTeamsTable = `CREATE TABLE teams (
			id serial4 PRIMARY KEY,
			name VARCHAR (50) NOT NULL,
			ident VARCHAR (50) UNIQUE NOT NULL,
			descr VARCHAR (255) NOT NULL
	);`

	// Creates the team-users relation table if not exists
	sqlCreateUsersTeamsTable = `CREATE TABLE users_teams (
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

	// Creates the data-policies table if not exists
	sqlCreateDataPoliciesTable = `CREATE TABLE data_policies (
		id serial4  PRIMARY KEY,
		descr VARCHAR (255) NOT NULL,
		use_aggregation bool NOT NULL,
		retention int NOT NULL,
		aggregation_retention int NOT NULL,
		aggregation_interval int NOT NULL
	);`
)
