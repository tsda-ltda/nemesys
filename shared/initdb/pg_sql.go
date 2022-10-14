package initdb

var sqlCommands []string = []string{
	// Users table
	`CREATE TABLE users (
		id serial4 PRIMARY KEY,
		name VARCHAR (50) NOT NULL,
		username VARCHAR (50) UNIQUE NOT NULL,
		password VARCHAR (255) NOT NULL,
		email VARCHAR (255) UNIQUE NOT NULL,
		role INT2 NOT NULL
	);`,

	// Teams table
	`CREATE TABLE teams (
		id serial4 PRIMARY KEY,
		name VARCHAR (50) NOT NULL,
		ident VARCHAR (50) UNIQUE NOT NULL,
		descr VARCHAR (255) NOT NULL
	);`,

	// Users teams table
	`CREATE TABLE users_teams (
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
	);`,

	// Data policy table
	`CREATE TABLE data_policies (
		id serial4  PRIMARY KEY,
		descr VARCHAR (255) NOT NULL,
		use_aggregation bool NOT NULL,
		retention int NOT NULL,
		aggregation_retention int NOT NULL,
		aggregation_interval int NOT NULL
	);`,

	// Context table
	`CREATE TABLE contexts (
		id serial4  PRIMARY KEY,
		teamId INTEGER NOT NULL,
		descr VARCHAR (255) NOT NULL,
		ident VARCHAR (50) NOT NULL,
		name VARCHAR (50) NOT NULL,
		CONSTRAINT fk_teamId
			FOREIGN KEY(teamId)
				REFERENCES teams(id)
				ON DELETE CASCADE
		
	);`,

	// Create context index
	`CREATE INDEX team_ident_index ON contexts (ident, teamId);`,
}
