// The following environment variables will be used to get a Postgres
// connection string: VB_TEST_DB.
//
package db

import (
	"context"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/stretchr/testify/assert"
	"github.com/veganbase/backend/chassis/test_utils"
	"github.com/veganbase/backend/services/api-gateway/db"
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

func TestCreateLoginToken(t *testing.T) {
	RunWithSchema(t, func(pg *db.PGClient, t *testing.T) {
		loadDefaultFixture(pg, t)

		// Create a token.
		rand.Seed(123)
		token1, err := pg.CreateLoginToken("test@example.com", "veganbase", "en")
		assert.Nil(t, err)

		// Force token collision.
		rand.Seed(123)
		token2, err := pg.CreateLoginToken("test2@example.com", "veganbase", "en")
		assert.Nil(t, err)
		assert.NotEqual(t, token1, token2)

		// Force token collision again.
		rand.Seed(123)
		token3, err := pg.CreateLoginToken("test3@example.com", "veganbase", "en")
		assert.Nil(t, err)
		assert.NotEqual(t, token1, token3)
		assert.NotEqual(t, token2, token3)

		// New token for user.
		tokenNew, err := pg.CreateLoginToken("test@example.com", "veganbase", "en")
		assert.Nil(t, err)
		assert.NotEqual(t, tokenNew, token1)
		assert.NotEqual(t, tokenNew, token2)
		assert.NotEqual(t, tokenNew, token3)
	})
}

func TestCheckLoginToken(t *testing.T) {
	RunWithSchema(t, func(pg *db.PGClient, t *testing.T) {
		loadDefaultFixture(pg, t)

		// Valid token.
		email, _, _, err := pg.CheckLoginToken("123456")
		assert.Nil(t, err)
		assert.Equal(t, email, "user@test.com")

		// Try to reuse valid token (should be single-use only).
		_, _, _, err = pg.CheckLoginToken("123456")
		assert.NotNil(t, err)

		// Invalid token.
		_, _, _, err = pg.CheckLoginToken("ABCDEF")
		assert.NotNil(t, err)
		assert.Equal(t, err, db.ErrLoginTokenNotFound)
	})
}

func TestSessions(t *testing.T) {
	RunWithSchema(t, func(pg *db.PGClient, t *testing.T) {
		loadDefaultFixture(pg, t)

		// Look up existing session.
		checkID, checkEmail, checkIsAdmin, err := pg.LookupSession("SESSION-1")
		assert.Nil(t, err)
		assert.Equal(t, checkID, "usr_TESTUSER1")
		assert.Equal(t, checkEmail, "test1@example.com")
		assert.Equal(t, checkIsAdmin, false)

		// Look up invalid session.
		checkID, checkEmail, checkIsAdmin, err = pg.LookupSession("SESSION-X")
		assert.NotNil(t, err)
		assert.Equal(t, err, db.ErrSessionNotFound)

		// Create and look up session.
		id, err := pg.CreateSession("usr_TESTUSER4", "user4@example.com", false)
		assert.Nil(t, err)
		checkID, checkEmail, checkIsAdmin, err = pg.LookupSession(id)
		assert.Nil(t, err)
		assert.Equal(t, checkID, "usr_TESTUSER4")
		assert.Equal(t, checkEmail, "user4@example.com")
		assert.Equal(t, checkIsAdmin, false)

		// Delete single session.
		_, _, _, err = pg.LookupSession("SESSION-3")
		assert.Nil(t, err)
		err = pg.DeleteSession("SESSION-3")
		assert.Nil(t, err)
		_, _, _, err = pg.LookupSession("SESSION-3")
		assert.NotNil(t, err)
		assert.Equal(t, err, db.ErrSessionNotFound)

		// Delete all sessions for a user.
		err = pg.DeleteUserSessions("usr_TESTUSER2")
		assert.Nil(t, err)
		_, _, _, err = pg.LookupSession("SESSION-2A")
		assert.NotNil(t, err)
		assert.Equal(t, err, db.ErrSessionNotFound)
		_, _, _, err = pg.LookupSession("SESSION-2B")
		assert.NotNil(t, err)
		assert.Equal(t, err, db.ErrSessionNotFound)
	})
}
