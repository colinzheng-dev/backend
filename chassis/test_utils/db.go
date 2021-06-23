package test_utils

import (
	"database/sql"
	"os"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

// InitTestDB does setup for the test database.
func InitTestDB(addPostgis bool) string {
	var err error
	pgdsn := os.Getenv("VB_TEST_DB")

	dbtmp, err := sqlx.Open("postgres", pgdsn)
	if err != nil {
		log.Fatal().Err(err).Msg("opening test database")
	}
	defer dbtmp.Close()
	ResetSchema(dbtmp, addPostgis)

	return pgdsn
}

// ResetSchema does a full pre-test schema reset, dropping all tables
// and creating a new empty schema.
func ResetSchema(db Execer, addPostgis bool) {
	MultiExec(db, dropSchema)
	if addPostgis {
		AddPostGIS(db)
	}
}

// AddPostGIS adds the PostGIS extension to a database.
func AddPostGIS(db Execer) {
	_, err := db.Exec("CREATE EXTENSION IF NOT EXISTS postgis")
	if err != nil {
		log.Fatal().Err(err).Msg("can't create postgis extension")
	}
}

var dropSchema = `
SET ROLE postgres;
DROP SCHEMA IF EXISTS public CASCADE;
CREATE SCHEMA IF NOT EXISTS public;
GRANT ALL ON SCHEMA public TO postgres;
GRANT ALL ON SCHEMA public TO public;
`

// Execer is an interface for database command execution.
type Execer interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

// MultiExec executes a script of SQL statements from a multiline
// string.
func MultiExec(db Execer, query string) {
	stmts := strings.Split(query, ";\n")
	if len(strings.Trim(stmts[len(stmts)-1], " \n\t\r")) == 0 {
		stmts = stmts[:len(stmts)-1]
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			log.Fatal().Msgf("executing '%s': %s", s, err.Error())
		}
	}
}
