package env

// All enviroment variables
var (
	// LOG LEVELS = (debug, info, warn, error, dpanic, panic, fatal)

	// ServiceManagerPingInterval is the interval of ping in seconds. Default is "5".
	ServiceManagerPingInterval = "5"

	// LogConsoleLevelAPIManager is the log level for console in API Manager service. Default is debug.
	LogConsoleLevelServiceManager = "debug"
	// LogBroadcastLevelAPIManager is the log level for broadcast in API Manager service. Default is info.
	LogBroadcastLevelServiceManager = "info"

	// LogConsoleLevelAPIManager is the log level for console in API Manager service. Default is debug.
	LogConsoleLevelAPIManager = "debug"
	// LogBroadcastLevelAPIManager is the log level for broadcast in API Manager service. Default is info.
	LogBroadcastLevelAPIManager = "info"

	// LogConsoleLevelSNMP is the log level for console in SNMP service. Default is debug.
	LogConsoleLevelSNMP = "debug"
	// LogBroadcastLevelSNMP is the log level for broadcast in SNMP service. Default is info.
	LogBroadcastLevelSNMP = "info"

	// LogConsoleLevelRTS is the log level for console in Real Time service. Default is debug.
	LogConsoleLevelRTS = "debug"
	// LogBroadcastLevelRTS is the log level for broadcast in Real Time service. Default is info.
	LogBroadcastLevelRTS = "info"

	// LogConsoleLevelDHS is the log level for console in Data History service. Default is debug.
	LogConsoleLevelDHS = "debug"
	// LogBroadcastLevelDHS is the log level for broadcast in Data History service. Default is info.
	LogBroadcastLevelDHS = "info"

	// LogConsoleLevelDHS is the log level for console in Data History service. Default is debug.
	LogConsoleLevelAlarmService = "debug"
	// LogBroadcastLevelDHS is the log level for broadcast in Data History service. Default is info.
	LogBroadcastLevelAlarmService = "info"

	// DefaultUsername is the default user's username. Default is "master".
	DefaultUsername = "master"
	// DefaultPassword is the default user's password (is strongly recommended to not use the default value). Default is "master".
	DefaultPassword = "master"

	// MaxDataPolicies is the max number of data policies. Default is "8".
	MaxDataPolicies = "8"

	// AMQPUsername is the amqp username. Default is "guest".
	AMQPUsername = "guest"
	// AMQPPassword is the amqp password. Default is "guest".
	AMQPPassword = "guest"
	// AMQPHost is the amqp host. Default is "localhost".
	AMQPHost = "localhost"
	// AMQPPort is the amqp port. Default is "5672".
	AMQPPort = "5672"

	// PGHost is the postgres host. Default is "127.0.0.1".
	PGHost = "127.0.0.1"
	// PGPort is the postgres port. Default is "5432".
	PGPort = "5432"
	// PGUsername is the postgres username. Default is "postgres".
	PGUsername = "postgres"
	// PGPassword is the postgres password. Default is "postgres".
	PGPassword = "postgres"
	// PGDBName is the database name. Default is "namesys".
	PGDBName = "namesys"
	// PGMaxConns is the postgres maximum number of connections open. Default is "20".
	PGMaxConns = "20"
	// PGMaxIdleConns is the postgres maximum number of idle connections. Default is "20".
	PGMaxIdleConns = "20"
	// PGConnMaxLifetime is the postgres maximum connection life time in minutes. Default is "10".
	PGConnMaxLifetime = "10"

	// InfluxHost is the influxdb host. Default is "localhost".
	InfluxHost = "localhost"
	// InfluxPort is the influxdb port. Default is "8086".
	InfluxPort = "8086"
	// InfluxOrg is the organization where data will be stored. Default is "nemesys".
	InfluxOrg = "nemesys"
	// InfluxToken is the influxdb token. Default is "".
	InfluxToken = ""
	// InfluxTLSCertFilePath is the path for the TLS certification file. Default is "".
	InfluxTLSCertFilePath = ""
	// InfluxTLSKeyFilePath is the path for the TLS key file. Default is "".
	InfluxTLSKeyFilePath = ""

	// DefaultLogsBucketRetention is the default retention in hours of the logs bucket. Default is "72".
	DefaultLogsBucketRetention = "72"

	// RDBAuthHost is redis for auth host. Default is "localhost".
	RDBAuthHost = "localhost"
	// RDBAuthPort is the redis for auth port. Default is "6379".
	RDBAuthPort = "6379"
	// RDBAuthDB is the redis for auth db. Default is "0".
	RDBAuthDB = "0"
	// RDBAuthPassword is the redis for auth password. Default is "".
	RDBAuthPassword = ""

	// RDBCacheHost is the redis for cache host. Default is "localhost".
	RDBCacheHost = "localhost"
	// RDBCachePort is the redis for cache port. Default is "6379".
	RDBCachePort = "6379"
	// RDBCacheDB is the redis for cache db. Default is "1".
	RDBCacheDB = "1"
	// RDBCachePassword is the redis for cache password. Default is "".
	RDBCachePassword = ""

	// APIManagerHost is the api manager host. Default is "localhost".
	APIManagerHost = "localhost"
	// APIManagerPort is the api manager port. Default is "9000".
	APIManagerPort = "9000"
	// APIManagerRoutesPrefix is the api manager routes prefix. Default is "api/v1".
	APIManagerRoutesPrefix = "api/v1"

	// UserSessionTTL is the user session TTL (time to live) in secods. Default is "604900" (one week).
	UserSessionTTL = "604800"
	// UserSessionTokenSize is the user session token size. Default is "64".
	UserSessionTokenSize = "64"
	// UserPWBcryptCost is the bcrypt cost for hashing users password. May vary according
	// to each machine config, the recommended is to set a cost
	// that makes '/login' route takes around 200ms. Default is "11".
	UserPWBcryptCost = "11"

	// DHSFlexLegacyDatalogWorkers is the number of flex-legacy datalog workers. Default is "3".
	DHSFlexLegacyDatalogWorkers = "3"

	// DHSFlexLegacyDatlogRequestInterval is the interval in minutes between each datalog request of a flex. Default is "60".
	DHSFlexLegacyDatlogRequestInterval = "60"

	// InicialDHSServices is the the number of inicial DHS services.
	InicialDHSServices = "1"

	// MetricAlarmEmailSender is the email of the sender. Default is "".
	MetricAlarmEmailSender = ""
	// MetricAlarmEmailSenderPassword is the password of the sender. Default is "".
	MetricAlarmEmailSenderPassword = ""
	// MetricAlarmEmailSenderHost is the host of the sender. Default is "";
	MetricAlarmEmailSenderHost = ""
	// MetricAlarmEmailSenderHostPort is the host port of the sender. Default is "";
	MetricAlarmEmailSenderHostPort = ""

	// AlarmServiceAMQPPublishers is the number of amqp publishers, which means number
	// of socket channels openned. Default is "1".
	AlarmServiceAMQPPublishers = "1"
	// APIManagerAMQPPublishers is the number of amqp publishers, which means number
	// of socket channels openned. Default is "3".
	APIManagerAMQPPublishers = "3"
	// DHSAMQPPublishers is the number of amqp publishers, which means number
	// of socket channels openned. Default is "5".
	DHSAMQPPublishers = "5"
	// RTSAMQPPublishers is the number of amqp publishers, which means number
	// of socket channels openned. Default is "5".
	RTSAMQPPublishers = "5"
	// SNMPAMQPPublishers is the number of amqp publishers, which means number
	// of socket channels openned. Default is "5".
	SNMPAMQPPublishers = "5"
)
