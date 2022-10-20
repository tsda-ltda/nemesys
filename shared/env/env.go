package env

// All enviroment variables
var (
	// LOG LEVELS = (debug, info, warn, error, dpanic, panic, fatal)
	// Minimum log level to enable console logs in API Manager service. Default is debug.
	LogConsoleLevelAPIManager = "debug"
	// Minimum log level to enable console logs in API Manager service. Default is info.
	LogBroadcastLevelAPIManager = "info"

	// Minimum log level to enable console logs in API Manager service. Default is debug.
	LogConsoleLevelSNMP = "debug"
	// Minimum log level to enable console logs in API Manager service. Default is info.
	LogBroadcastLevelSNMP = "info"

	// Default master username. Default is "master".
	Username = "master"
	// Default master password (is strongly recommended to not use the default value). Default is "admin".
	PW = "master"

	// Max data policies. Default is "8".
	MaxDataPolicies = "8"

	// AMQP Username. Default is "guest".
	AMQPUsername = "guest"
	// AMQP password. Default is "guest".
	AMQPPassword = "guest"
	// AMQP host. Default is "localhost".
	AMQPHost = "localhost"
	// AMQP port. Defaul is "5672".
	AMQPPort = "5672"

	// Postgresql host. Default is "127.0.0.1".
	PGHost = "127.0.0.1"
	// Postgresql port. Default is "5432".
	PGPort = "5432"
	// Postgresql username. Default is "postgres".
	PGUsername = "postgres"
	// Postgresql password. Default is "postgres".
	PGPW = "postgres"
	// Postgresql database name. Default is "dev".
	PGDBName = "dev"

	// Redis for authentication host. Default is "localhost".
	RDBAuthHost = "localhost"
	// Redis for authentication host. Default is "6379".
	RDBAuthPort = "6379"
	// Redis for authentication database. Default is "0".
	RDBAuthDB = "0"
	// Redis for authentication password. Default is "".
	RDBAuthPW = ""

	// API Manager host. Default is "localhost".
	APIManagerHost = "localhost"
	// API Manager port. Default is "9000".
	APIManagerPort = "9000"
	// API Manager port. Default is "9000".
	APIManagerRoutesPrefix = "/api/v1"

	// User session cookie time to live (seconds). Default is "604900" (one week).
	UserSessionTTL = "604800"
	// User session token character size. Default is "64".
	UserSessionTokenSize = "64"
	// The bcrypt cost for hashing users password. May vary according
	// to each machine config, the recommended is to set a cost
	// that makes '/login' route take around 200ms. Default is "11".
	UserPWBcryptCost = "11"
)
