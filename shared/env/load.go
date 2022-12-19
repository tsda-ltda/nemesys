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
	set("LOG_CONSOLE_LEVEL_SERVICE_MANAGER", &LogConsoleLevelServiceManager)
	set("LOG_BROADCAST_LEVEL_SERVICE_MANAGER", &LogBroadcastLevelServiceManager)

	set("LOG_CONSOLE_LEVEL_API_MANAGER", &LogConsoleLevelAPIManager)
	set("LOG_BROADCAST_LEVEL_API_MANAGER", &LogBroadcastLevelAPIManager)

	set("LOG_CONSOLE_LEVEL_SNMP", &LogConsoleLevelSNMP)
	set("LOG_BROADCAST_LEVEL_SNMP", &LogBroadcastLevelSNMP)

	set("LOG_CONSOLE_LEVEL_RTS", &LogConsoleLevelRTS)
	set("LOG_BROADCAST_LEVEL_RTS", &LogBroadcastLevelRTS)

	set("LOG_CONSOLE_LEVEL_DHS", &LogConsoleLevelDHS)
	set("LOG_BROADCAST_LEVEL_DHS", &LogBroadcastLevelDHS)

	set("LOG_CONSOLE_LEVEL_ALARM_SERVICE", &LogConsoleLevelAlarmService)
	set("LOG_BROADCAST_LEVEL_ALARM_SERVICE", &LogBroadcastLevelAlarmService)

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
	set("PG_MAX_CONN_LIFETIME", &PGConnMaxLifetime)
	set("PG_MAX_CONNS", &PGMaxConns)
	set("PG_MAX_IDLE_CONNS", &PGMaxIdleConns)

	set("INFLUX_HOST", &InfluxHost)
	set("INFLUX_PORT", &InfluxPort)
	set("INFLUX_ORG", &InfluxOrg)
	set("INFLUX_TOKEN", &InfluxToken)
	set("INFLUX_TLS_CERT_FILE_PATH", &InfluxTLSCertFilePath)
	set("INFLUX_TLS_KEY_FILE_PATH", &InfluxTLSKeyFilePath)

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

	set("DHS_FLEX_LEGACY_DATALOG_WORKERS", &DHSFlexLegacyDatalogWorkers)
	set("DHS_FLEX_LEGACY_DATALOG_REQUEST_INTERVAL", &DHSFlexLegacyDatlogRequestInterval)
	set("INICIAL_DHS_SERVICES", &InicialDHSServices)

	set("METRIC_ALARM_EMAIL_SENDER", &MetricAlarmEmailSender)
	set("METRIC_ALARM_EMAIL_SENDER_PASSWORD", &MetricAlarmEmailSenderPassword)
	set("METRIC_ALARM_EMAIL_SENDER_HOST", &MetricAlarmEmailSenderHost)
	set("METRIC_ALARM_EMAIL_SENDER_HOST_PORT", &MetricAlarmEmailSenderHostPort)

	set("ALARM_HISTORY_BUCKET_RETENTION", &AlarmHistoryBucketRetention)
	set("REQUESTS_COUNT_BUCKET_RETENTION", &RequestsCountBucketRetention)
	set("LOGS_BUCKET_RETENTION", &LogsBucketRetention)

	set("ALARM_SERVICE_AMQP_PUBLISHERS", &AlarmServiceAMQPPublishers)
	set("API_MANAGER_AMQP_PUBLISHERS", &APIManagerAMQPPublishers)
	set("DHS_SERVICE_AMQP_PUBLISHERS", &DHSAMQPPublishers)
	set("RTS_SERVICE_AMQP_PUBLISHERS", &RTSAMQPPublishers)
	set("SNMP_AMQP_PUBLISHERS", &SNMPAMQPPublishers)
}
