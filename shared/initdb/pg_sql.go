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

	// API Keys table
	`CREATE TABLE apikeys (
		id SERIAL4 PRIMARY KEY,
		user_id INT4 NOT NULL,
		descr VARCHAR(255) NOT NULL, 
		ttl INT2 NOT NULL,
		created_at INT8 NOT NULL,
		CONSTRAINT apikeys_fk_user_id
		FOREIGN KEY(user_id)
				REFERENCES users(id)
				ON DELETE CASCADE
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
		CONSTRAINT users_teams_fk_user_id
		FOREIGN KEY(user_id)
				REFERENCES users(id)
				ON DELETE CASCADE,
		CONSTRAINT users_teams_fk_team_id
			FOREIGN KEY(team_id)
				REFERENCES teams(id)
				ON DELETE CASCADE
	);`,

	// Data policy table
	`CREATE TABLE data_policies (
		id SERIAL2  PRIMARY KEY,
		descr VARCHAR (255) NOT NULL,
		retention INT4 NOT NULL,
		use_aggr BOOLEAN NOT NULL,
		aggr_retention INT4 NOT NULL,
		aggr_interval INT4 NOT NULL,
		aggr_fn VARCHAR(50) NOT NULL
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
		CONSTRAINT metrics_fk_container_id
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
		CONSTRAINT metrics_ref_fk_metric_id
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
		CONSTRAINT snmpv2c_containers_fk_container_id
			FOREIGN KEY(container_id)
				REFERENCES containers(id)
				ON DELETE CASCADE
				DEFERRABLE INITIALLY DEFERRED
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
		serial_number INT4 UNIQUE NOT NULL,
		model INT2 NOT NULL,
		city VARCHAR (50) NOT NULL,
		region VARCHAR (50) NOT NULL,
		country VARCHAR (50) NOT NULL,
		CONSTRAINT flex_legacy_containers_fk_container_id
			FOREIGN KEY(container_id)
				REFERENCES containers(id)
				ON DELETE CASCADE
				DEFERRABLE INITIALLY DEFERRED
	);`,

	// Create index on target and port
	`CREATE UNIQUE INDEX flex_legacy_target_port_index ON flex_legacy_containers (target, port);`,

	// Create flex legacy datalog download registry
	`CREATE TABLE flex_legacy_datalog_download_registry (
		container_id INT4 UNIQUE NOT NULL,
		metering INT8 NOT NULL,
		status INT8 NOT NULL,
		command INT8 NOT NULL,
		virtual INT8 NOT NULL,
		CONSTRAINT flex_legacy_datalog_download_registry_fk_container_id
			FOREIGN KEY(container_id)
				REFERENCES containers(id)
				ON DELETE CASCADE
	);`,

	// SNMP metrics table
	`CREATE TABLE snmpv2c_metrics (
		metric_id INT8 UNIQUE NOT NULL,
		oid VARCHAR (128) NOT NULL,
		CONSTRAINT snmpv2c_metrics_fk_metric_id
			FOREIGN KEY(metric_id)
				REFERENCES metrics(id)
				ON DELETE CASCADE
				DEFERRABLE INITIALLY DEFERRED
	);`,

	// Flex Legacy metrics table
	`CREATE TABLE flex_legacy_metrics (
		metric_id INT8 UNIQUE NOT NULL,
		oid VARCHAR (128) NOT NULL,
		port INT2 NOT NULL,
		port_type INT2 NOT NULL,
		CONSTRAINT flex_legacy_metrics_fk_metric_id
			FOREIGN KEY(metric_id)
				REFERENCES metrics(id)
				ON DELETE CASCADE
				DEFERRABLE INITIALLY DEFERRED
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
		CONSTRAINT contextual_metrics_fk_ctx_id
			FOREIGN KEY(ctx_id)
				REFERENCES contexts(id)
				ON DELETE CASCADE,
		CONSTRAINT contextual_metrics_fk_metric_id
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

	// Create alarm expression table
	`CREATE TABLE alarm_expressions (
		metric_id INT8 UNIQUE NOT NULL,
		minor_expression VARCHAR (255) NOT NULL,
		major_expression VARCHAR (255) NOT NULL,
		critical_expression VARCHAR (255) NOT NULL,
		minor_descr VARCHAR (255) NOT NULL,
		major_descr VARCHAR (255) NOT NULL,
		critical_descr VARCHAR (255) NOT NULL,
		CONSTRAINT alarm_expressions_fk_metric_id
			FOREIGN KEY(metric_id)
				REFERENCES metrics(id)
				ON DELETE CASCADE
	);`,

	// Create alarm profile table
	`CREATE TABLE alarm_profiles (
		id SERIAL4 PRIMARY KEY,
		name VARCHAR (50) NOT NULL,
		minor BOOLEAN NOT NULL,
		major BOOLEAN NOT NULL,
		critical BOOLEAN NOT NULL
	);`,

	// Create alarm profile relation table
	`CREATE TABLE alarm_profiles_rel (
		alarm_profile_id INT4 NOT NULL,
		metric_id INT8 NOT NULL,
		CONSTRAINT alarm_profiles_rel_fk_metric_id
			FOREIGN KEY(metric_id)
				REFERENCES metrics(id)
				ON DELETE CASCADE,
		CONSTRAINT alarm_profiles_rel_fk_alarm_profile_id
			FOREIGN KEY(alarm_profile_id)
				REFERENCES alarm_profiles(id)
				ON DELETE CASCADE
	);`,

	// Create metric alarm state table
	`CREATE TABLE alarm_states (
		metric_id INT8 UNIQUE NOT NULL,
		state INT2 NOT NULL,
		last_minor_time INT8 NOT NULL,
		last_major_time INT8 NOT NULL,
		last_critical_time INT8 NOT NULL,
		last_recognization_time INT8 NOT NULL,
		always_alarmed_on_new_alarm BOOLEAN NOT NULL,
		recognization_max_lifetime INT8 NOT NULL	
	);`,

	// Create request registry table
	`CREATE TABLE request_registry (
		requests INT8,
		realtime_data_requests INT8,
		history_data_requests INT8
	);`,
	`INSERT INTO request_registry (requests, realtime_data_requests, history_data_requests) VALUES (0,0,0);`,

	// Create server cost price table
	`CREATE TABLE price_table (
		id INT4 PRIMARY KEY,
		coin_type VARCHAR (5) NOT NULL, 
		_user FLOAT8 NOT NULL,
		team FLOAT8 NOT NULL,
		context FLOAT8 NOT NULL,
		contextual_metric FLOAT8 NOT NULL,
		basic_container FLOAT8 NOT NULL,
		snmpv2c_container FLOAT8 NOT NULL,
		flex_legacy_container FLOAT8 NOT NULL,
		basic_metric FLOAT8 NOT NULL,
		snmpv2c_metric FLOAT8 NOT NULL,
		flex_legacy_metric FLOAT8 NOT NULL,
		custom_query FLOAT8 NOT NULL,
		data_policy FLOAT8 NOT NULL,
		alarm_expression FLOAT8 NOT NULL,
		alarm_profile FLOAT8 NOT NULL,
		refkey FLOAT8 NOT NULL,
		api_key FLOAT8 NOT NULL,
		influx_data_point FLOAT8 NOT NULL,
		request FLOAT8 NOT NULL,
		realtime_data_request FLOAT8 NOT NULL,
		history_data_request FLOAT8 NOT NULL
	);`,
	`INSERT INTO price_table (id, coin_type, _user, team, context, contextual_metric, basic_container,
		snmpv2c_container, flex_legacy_container, basic_metric, snmpv2c_metric, flex_legacy_metric, custom_query,
		data_policy, alarm_expression, alarm_profile, refkey, api_key, influx_data_point, request, realtime_data_request,
		history_data_request) VALUES (1, 'R$', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)`,

	// Create base plan table
	`CREATE TABLE base_plan (
		id INT4 PRIMARY KEY,
		cost FLOAT8 NOT NULL,
		users INT4 NOT NULL,
		teams INT4 NOT NULL,
		contexts INT4 NOT NULL,
		contextual_metrics INT4 NOT NULL,
		basic_containers INT4 NOT NULL,
		snmpv2c_containers INT4 NOT NULL,
		flex_legacy_containers INT4 NOT NULL,
		basic_metrics INT8 NOT NULL,
		snmpv2c_metrics INT8 NOT NULL,
		flex_legacy_metrics INT8 NOT NULL,
		custom_queries INT4 NOT NULL,
		data_policies INT4 NOT NULL,
		alarm_expressions INT4 NOT NULL,
		alarm_profiles INT4 NOT NULL,
		refkeys INT8 NOT NULL,
		api_keys INT4 NOT NULL,
		influx_data_points INT8 NOT NULL,
		requests INT8 NOT NULL,
		realtime_data_requests INT8 NOT NULL,
		history_data_requests INT8 NOT NULL
	);`,
	`INSERT INTO base_plan (id, cost, users, teams, contexts, contextual_metrics, basic_containers,
		snmpv2c_containers, flex_legacy_containers, basic_metrics, snmpv2c_metrics, flex_legacy_metrics, custom_queries,
		data_policies, alarm_expressions, alarm_profiles, refkeys, api_keys, influx_data_points, requests, realtime_data_requests,
		history_data_requests) VALUES (1,0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)`,

	// Create requests users whitelist
	`CREATE TABLE counter_whitelist (
		user_id INT4 NOT NULL,
		CONSTRAINT requests_counter_whitelist_fk_user_id
			FOREIGN KEY(user_id)
				REFERENCES users(id)
				ON DELETE CASCADE
	);`,
}
