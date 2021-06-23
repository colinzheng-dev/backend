// The following environment variables will be used to get a Postgres
// connection string: VB_TEST_DB.
//
package db

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/stretchr/testify/assert"

	"github.com/veganbase/backend/chassis/test_utils"
	"github.com/veganbase/backend/services/search-service/db"
)

var pg *db.PGClient

func init() {
	pgdsn := test_utils.InitTestDB(true)
	var err error
	pg, err = db.NewPGClient(context.Background(), pgdsn)
	if err != nil {
		log.Fatal().Err(err).Msg("couldn't connect to test database")
	}
}

func RunWithSchema(t *testing.T, test func(pg *db.PGClient, t *testing.T)) {
	defer func() {
		test_utils.ResetSchema(pg.DB, true)
	}()

	migrations := &migrate.AssetMigrationSource{
		Asset:    db.Asset,
		AssetDir: db.AssetDir,
		Dir:      "migrations",
	}
	_, err := migrate.Exec(pg.DB.DB, "postgres", migrations, migrate.Up)
	assert.Nil(t, err, "database migrations failed!", err)

	test(pg, t)
}

func loadDefaultFixture(db *db.PGClient, t *testing.T) {
	f, err := os.Open("fixture.sql")
	assert.Nil(t, err)
	defer f.Close()
	fixture, err := ioutil.ReadAll(f)
	assert.Nil(t, err)
	tx := pg.DB.MustBegin()
	test_utils.MultiExec(tx, string(fixture))
	tx.Commit()
}

func TestFullText(t *testing.T) {
	RunWithSchema(t, func(pg *db.PGClient, t *testing.T) {
		loadDefaultFixture(pg, t)

		var tests = []struct {
			query string
			ids   []string
		}{
			{"cream tea", []string{"item0002", "item0004"}},
			{"hotel", []string{"item0001", "item0003", "item0002"}},
			{"B&B", []string{"item0004"}},
			{"restaurant", []string{}},
		}
		for _, test := range tests {
			ids, err := pg.FullText(test.query, nil, nil)
			assert.Nil(t, err)
			assert.NotNil(t, ids)
			fmt.Println(ids)
			assert.ElementsMatch(t, ids, test.ids)
		}
	})
}

func TestGeo(t *testing.T) {
	RunWithSchema(t, func(pg *db.PGClient, t *testing.T) {
		loadDefaultFixture(pg, t)

		var tests = []struct {
			latitude  float64
			longitude float64
			dist      float64
			ids       []string
		}{
			{51.507, -0.144, 10, []string{"item0002"}},
			{49.5, -2.2, 500, []string{"item0004", "item0002"}},
		}
		for _, test := range tests {
			ids, err := pg.Geo(test.latitude, test.longitude, test.dist, nil, nil)
			assert.Nil(t, err)
			assert.NotNil(t, ids)
			fmt.Println(ids)
			assert.ElementsMatch(t, ids, test.ids)
		}
	})
}
