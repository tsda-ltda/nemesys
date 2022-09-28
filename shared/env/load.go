package env

import (
	"log"
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

// Init initialize all variables in the package according to the enviroment.
func Init() {
	// get env file path
	path := load("ENV_FILE", ".env")

	// load env file
	err := godotenv.Load(path)
	if err != nil {
		log.Printf("\nfail to load enviroment file, path: %s, err: %s", path, err)
	}

	// set config
	set("PG_HOST", &PGHost)
	set("PG_PORT", &PGPort)
	set("PG_USERNAME", &PGUsername)
	set("PG_PW", &PGPW)
	set("PG_DB_NAME", &PGDBName)

	set("RDB_AUTH_HOST", &RDBAuthHost)
	set("RDB_AUTH_PORT", &RDBAuthPort)
	set("RDB_AUTH_DB", &RDBAuthDB)
	set("RDB_AUTH_PW", &RDBAuthPW)

}
