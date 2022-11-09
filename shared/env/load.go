package env

import (
	"os"

	"github.com/joho/godotenv"
)

// load loads the selected key in enviroment, if key value equal to
// "", returns defaultValue, otherwise returns env value.
func load(key string, defaultValue string) string {
	e := os.Getenv(key)
	if len(e) == 0 {
		return defaultValue
	}
	return e
}

// set get the selected key in enviroment, if key value different then
// "", sets in out value.
func set(key string, out *string) {
	e := os.Getenv(key)
	if len(e) == 0 {
		return
	}
	*out = e
}

func LoadEnvFile() error {
	// get env file path
	path := load("ENV_FILE", ".env")

	// load env file
	return godotenv.Load(path)
}

// Init initialize all variables in the package according to the enviroment.
func Init() {
	// set config
	set("LOG_CONSOLE_LEVEL_API_MANAGER", &LogConsoleLevelAPIManager)
	set("LOG_BROADCAST_LEVEL_API_MANAGER", &LogBroadcastLevelAPIManager)

	set("LOG_CONSOLE_LEVEL_SNMP", &LogConsoleLevelSNMP)
	set("LOG_BROADCAST_LEVEL_SNMP", &LogBroadcastLevelSNMP)

	set("LOG_CONSOLE_LEVEL_RTS", &LogConsoleLevelRTS)
	set("LOG_BROADCAST_LEVEL_RTS", &LogBroadcastLevelRTS)

	set("DEFAULT_USERNAME", &DefaultUsername)
	set("DEFAULT_PASSWORD", &DefaultPassword)

	set("MAX_DATA_POLICIES", &MaxDataPolicies)

	set("AMQP_USERNAME", &AMQPUsername)
	set("AMQP_PASSWORD", &AMQPPassword)
	set("AMQP_PORT", &AMQPPort)
	set("AMQP_HOST", &AMQPHost)

	set("PG_HOST", &PGHost)
	set("PG_PORT", &PGPort)
	set("PG_USERNAME", &PGUsername)
	set("PG_PASSWORD", &PGPassword)
	set("PG_DB_NAME", &PGDBName)

	set("RDB_AUTH_HOST", &RDBAuthHost)
	set("RDB_AUTH_PORT", &RDBAuthPort)
	set("RDB_AUTH_DB", &RDBAuthDB)
	set("RDB_AUTH_PASSWORD", &RDBAuthPassword)

	set("RDB_CACHE_HOST", &RDBCacheHost)
	set("RDB_CACHE_PORT", &RDBCachePort)
	set("RDB_CACHE_DB", &RDBCacheDB)
	set("RDB_CACHE_PASSWORD", &RDBCachePassword)

	set("API_MANAGER_HOST", &APIManagerHost)
	set("API_MANAGER_PORT", &APIManagerPort)
	set("API_MANAGER_ROUTES_PREFIX", &APIManagerRoutesPrefix)

	set("USER_SESSION_TTL", &UserSessionTTL)
	set("USER_SESSION_TOKEN_SIZE", &UserSessionTokenSize)
	set("USER_PW_BCRYPT_COST", &UserPWBcryptCost)
}
