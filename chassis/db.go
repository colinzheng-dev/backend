package chassis

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	migrate "github.com/rubenv/sql-migrate"
)

type assetFunc func(name string) ([]byte, error)
type assetDirFunc func(name string) ([]string, error)

// DBConnect creates a new database connection.
func DBConnect(ctx context.Context, dbName string, dbURL string,
	asset assetFunc, assetDir assetDirFunc) (*sqlx.DB, error) {
	// Connect to database and test connection integrity.
	db, err := sqlx.Open("postgres", dbURL)
	if err != nil {
		return nil, errors.Wrap(err, "opening "+dbName+" database")
	}
	err = db.Ping()
	if err != nil {
		return nil, errors.Wrap(err, "pinging "+dbName+" database")
	}

	// Limit maximum connections (default is unlimited).
	db.SetMaxOpenConns(10)

	// Run and log database migrations.
	migrations := &migrate.AssetMigrationSource{
		Asset:    asset,
		AssetDir: assetDir,
		Dir:      "migrations",
	}
	n, err := migrate.Exec(db.DB, "postgres", migrations, migrate.Up)
	if err != nil {
		log.Fatal().Err(err).Msgf("database migrations failed! %s", err)
	}
	log.Info().Msgf("applied new database migrations: %d", n)
	migrationRecords, err := migrate.GetMigrationRecords(db.DB, "postgres")
	if err != nil {
		log.Fatal().Err(err).Msg("couldn't read back database migration records")
	}
	if len(migrationRecords) == 0 {
		log.Info().Msg("no database migrations currently applied")
	} else {
		for _, m := range migrationRecords {
			log.Info().
				Str("migration", m.Id).
				Time("applied_at", m.AppliedAt).
				Msg("database migration")
		}
	}

	return db, nil
}

// Paginate does the API-wide processing of pagination controls:
// maximum page size is 100, default page size is 30, default page is
// 1.
func Paginate(page, perPage uint) string {
	limit := uint(30)
	if perPage != 0 {
		limit = perPage
		if limit > 100 {
			limit = uint(100)
		}
	}
	offset := uint(0)
	if page != 0 {
		offset = (page - 1) * limit
	}
	return fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)
}

// Tags is a type wrapper around an array of tags. To make tag
// searches quick, the tags are stored in a JSONB column as a JSON
// object, i.e. the tag list "cake,chocolate" is represented in JSON
// as '{"cake":true,"chocolate":true}'. This means that some special
// marshalling is needed for database access, handled by the Value and
// Scan functions.
type Tags []string

// Value encodes a tags value for the database.
func (tags Tags) Value() (driver.Value, error) {
	tagmap := map[string]bool{}
	for _, t := range tags {
		tagmap[t] = true
	}
	return json.Marshal(tagmap)
}

// Scan decodes a tags value from the database.
func (tags *Tags) Scan(value interface{}) error {
	var b []byte
	switch v := value.(type) {
	case string:
		b = []byte(v)
	case []byte:
		b = v
	default:
		return errors.New("incompatible type for Tags")
	}
	tagmap := map[string]bool{}
	err := json.Unmarshal([]byte(b), &tagmap)
	if err != nil {
		return errors.New("tags can't be decoded from JSON")
	}
	tagtmp := []string{}
	for k := range tagmap {
		tagtmp = append(tagtmp, k)
	}
	*tags = Tags(tagtmp)
	return nil
}
