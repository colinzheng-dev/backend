// The following environment variables will be used to get a Postgres
// connection string: VB_TEST_DB.
//
package db

import (
	"context"
	"testing"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/stretchr/testify/assert"

	"github.com/veganbase/backend/chassis/test_utils"
	"github.com/veganbase/backend/services/category-service/db"
)

var pg *db.PGClient

func init() {
	pgdsn := test_utils.InitTestDB(false)
	var err error
	pg, err = db.NewPGClient(context.Background(), pgdsn)
	if err != nil {
		log.Fatal().Err(err).Msg("couldn't connect to test database")
	}
}

func RunWithSchema(t *testing.T, test func(pg *db.PGClient, t *testing.T)) {
	defer func() {
		test_utils.ResetSchema(pg.DB, false)
	}()

	migrations := &migrate.AssetMigrationSource{
		Asset:    db.Asset,
		AssetDir: db.AssetDir,
		Dir:      "migrations",
	}
	_, err := migrate.Exec(pg.DB.DB, "postgres", migrations, migrate.Up)
	assert.Nil(t, err, "database migrations failed!")

	test(pg, t)
}

func TestCategoryList(t *testing.T) {
	RunWithSchema(t, func(pg *db.PGClient, t *testing.T) {
		cats, err := pg.Categories()
		assert.Nil(t, err)

		assert.Contains(t, cats, "room-type")
		assert.Equal(t, cats["room-type"].Schema.String(), `{"type": "string"}`)
		assert.Equal(t, cats["room-type"].Label, "Room types")
		assert.False(t, cats["room-type"].Extensible)

		assert.Contains(t, cats, "nutrition-info")
		assert.True(t, cats["nutrition-info"].Extensible)
	})
}

func TestCategoryRetrieval(t *testing.T) {
	RunWithSchema(t, func(pg *db.PGClient, t *testing.T) {
		cat, err := pg.CategoryEntries("doesnt-exist", nil)
		assert.Nil(t, cat)
		assert.Equal(t, err, db.ErrCategoryNotFound)

		cat, err = pg.CategoryEntries("room-type", nil)
		assert.Nil(t, err)
		assert.Contains(t, cat, "single")
		s1, ok := cat["single"]
		assert.True(t, ok)
		assert.Equal(t, s1, "single")

		cat, err = pg.CategoryEntries("nutrition-info", nil)
		assert.Nil(t, err)
		assert.Contains(t, cat, "vitamin-a")
		s2, ok := cat["vitamin-a"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, s2["name"], "Vitamin A")
	})
}
