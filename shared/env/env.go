package env

// All enviroment variables
var (
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

	// User session cookie time to live (seconds). Default is "604900" (one week).
	UserSessionTTL = "604800"
	// User session token character size. Default is "64".
	UserSessionTokenSize = "64"
	// The bcrypt cost for hashing users password. May vary according
	// to each machine config, the recommended is to set a cost
	// that makes '/login' route take around 200ms. Default is "11".
	UserPWBcryptCost = "11"
)
