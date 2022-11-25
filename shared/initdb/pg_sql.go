package initdb

var sqlCommands []string = []string{
	// Users table
	`CREATE TABLE users (
		id SERIAL4 PRIMARY KEY,
		name VARCHAR (50) NOT NULL,
		username VARCHAR (50) UNIQUE NOT NULL,
		password VARCHAR (255) NOT NULL,
		email VARCHAR (255) UNIQUE NOT NULL,
		role INT2 NOT NULL
	);`,

	// Teams table
	`CREATE TABLE teams (
		id SERIAL4 PRIMARY KEY,
		name VARCHAR (50) NOT NULL,
		ident VARCHAR (50) UNIQUE NOT NULL,
		descr VARCHAR (255) NOT NULL
	);`,

	// Users teams table
	`CREATE TABLE users_teams (
			user_id INT4,
		team_id INT4,
		CONSTRAINT fk_user_id
		FOREIGN KEY(user_id)
				REFERENCES users(id)
				ON DELETE CASCADE,
		CONSTRAINT fk_team_id
			FOREIGN KEY(team_id)
				REFERENCES teams(id)
				ON DELETE CASCADE
	);`,

	// Data policy table
	`CREATE TABLE data_policies (
		id SERIAL2  PRIMARY KEY,
		descr VARCHAR (255) NOT NULL,
		use_aggregation BOOLEAN NOT NULL,
		retention INT4 NOT NULL,
		aggregation_retention INT4 NOT NULL,
		aggregation_interval INT4 NOT NULL
	);`,

	// Containers table
	`CREATE TABLE containers (
		id SERIAL4 PRIMARY KEY,
		name VARCHAR (50) NOT NULL,
		descr VARCHAR (255) NOT NULL,
		type INT2 NOT NULL,
		enabled BOOLEAN NOT NULL,
		rts_pulling_interval INT4 NOT NULL
	);`,

	// Create container index
	`CREATE INDEX container_type_index ON containers (type);`,

	// Metrics table
	`CREATE TABLE metrics (
		id SERIAL8 PRIMARY KEY,
		container_id INT4 NOT NULL,
		container_type INT2 NOT NULL,
		type INT2 NOT NULL,
		name VARCHAR (50) NOT NULL,
		descr VARCHAR (255) NOT NULL,
		enabled BOOLEAN NOT NULL,
		data_policy_id INT4 NOT NULL,
		rts_pulling_times INT2 NOT NULL,
		rts_data_cache_duration INT4 NOT NULL,
		dhs_enabled BOOLEAN NOT NULL,
		dhs_interval INT4 NOT NULL,
		ev_expression VARCHAR (255) NOT NULL,
		CONSTRAINT fk_container_id
			FOREIGN KEY(container_id)
				REFERENCES containers(id)
				ON DELETE CASCADE
	);`,

	// Create metric index
	`CREATE INDEX metrics_container_id_index ON metrics (container_id);`,
	`CREATE INDEX metrics_container_type_index ON metrics (container_type);`,

	// Create metrics ref table
	`CREATE TABLE metrics_ref (
		id SERIAL8 PRIMARY KEY,
		refkey VARCHAR (200) UNIQUE NOT NULL,
		metric_id INT8 NOT NULL,
		CONSTRAINT fk_metric_id
			FOREIGN KEY(metric_id)
				REFERENCES metrics(id)
				ON DELETE CASCADE
	);`,

	// Create metrics ref index
	`CREATE INDEX metrics_ref_metric_id ON metrics_ref (metric_id);`,

	// SNMP Container table
	`CREATE TABLE snmpv2c_containers (
		container_id INT4 UNIQUE NOT NULL,
		target VARCHAR (15) NOT NULL,
		port INT4 NOT NULL,
		transport VARCHAR (3) NOT NULL,
		community VARCHAR (50) NOT NULL,
		retries INT2 NOT NULL,
		max_oids INT2 NOT NULL,
		timeout INT4 NOT NULL,
		CONSTRAINT fk_container_id
			FOREIGN KEY(container_id)
				REFERENCES containers(id)
				ON DELETE CASCADE
	);`,

	// Create index on target and port
	`CREATE UNIQUE INDEX snmpv2c_containers_target_port_index ON snmpv2c_containers (target, port);`,

	// Create Flex Legacy container
	`CREATE TABLE flex_legacy_containers (
		container_id INT4 UNIQUE NOT NULL,
		target VARCHAR (15) NOT NULL,
		port INT4 NOT NULL,
		transport VARCHAR (3) NOT NULL,
		community VARCHAR (50) NOT NULL,
		retries INT2 NOT NULL,
		max_oids INT2 NOT NULL,
		timeout INT4 NOT NULL,
		cache_duration INT4 NOT NULL,
		serial_number INT4 UNIQUE NOT NULL,
		model INT2 NOT NULL,
		city VARCHAR (50) NOT NULL,
		region VARCHAR (50) NOT NULL,
		country VARCHAR (50) NOT NULL,
		CONSTRAINT fk_container_id
			FOREIGN KEY(container_id)
				REFERENCES containers(id)
				ON DELETE CASCADE
	);`,

	// Create index on target and port
	`CREATE UNIQUE INDEX flex_legacy_target_port_index ON flex_legacy_containers (target, port);`,

	// SNMP metrics table
	`CREATE TABLE snmpv2c_metrics (
		metric_id INT8 UNIQUE NOT NULL,
		oid VARCHAR (128) NOT NULL,
		CONSTRAINT fk_metric_id
			FOREIGN KEY(metric_id)
				REFERENCES metrics(id)
				ON DELETE CASCADE
	);`,

	// Flex Legacy metrics table
	`CREATE TABLE flex_legacy_metrics (
		metric_id INT8 UNIQUE NOT NULL,
		oid VARCHAR (128) NOT NULL,
		port INT2 NOT NULL,
		port_type INT2 NOT NULL,
		CONSTRAINT fk_metric_id
			FOREIGN KEY(metric_id)
				REFERENCES metrics(id)
				ON DELETE CASCADE
	);`,

	// Context table
	`CREATE TABLE contexts (
		id SERIAL4 PRIMARY KEY,
		team_id INT4 NOT NULL,
		ident VARCHAR (50) NOT NULL,
		name VARCHAR (50) NOT NULL,
		descr VARCHAR (255) NOT NULL,
		CONSTRAINT fk_team_id
			FOREIGN KEY(team_id)
				REFERENCES teams(id)
				ON DELETE CASCADE
	);`,

	// Create context index
	`CREATE INDEX context_team_ident_index ON contexts (ident, team_id);`,

	// Create contextual metrics
	`CREATE TABLE contextual_metrics (
		id SERIAL8 PRIMARY KEY,
		ctx_id INT4 NOT NULL,
		metric_id INT8 NOT NULL,
		ident VARCHAR (50) NOT NULL,
		name VARCHAR (50) NOT NULL,
		descr VARCHAR (255) NOT NULL,
		CONSTRAINT fk_ctx_id
			FOREIGN KEY(ctx_id)
				REFERENCES contexts(id)
				ON DELETE CASCADE,
		CONSTRAINT fk_metric_id
			FOREIGN KEY(metric_id)
				REFERENCES metrics(id)
				ON DELETE CASCADE
	);`,

	// Create contextual metrics
	`CREATE INDEX contextual_metric_ctx_id ON contextual_metrics (ctx_id);`,
	`CREATE INDEX contextual_metric_ident_id ON contextual_metrics (ident);`,

	// Create custom queries table
	`CREATE TABLE custom_queries (
		id SERIAL4 PRIMARY KEY,
		ident VARCHAR (50) NOT NULL UNIQUE,
		descr VARCHAR (255) NOT NULL,
		flux VARCHAR (1000) NOT NULL
	);`,
}
