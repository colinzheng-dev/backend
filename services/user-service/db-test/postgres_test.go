// The following environment variables will be used to get a Postgres
// connection string: VB_TEST_DB.
//
package db

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/stretchr/testify/assert"

	"github.com/veganbase/backend/chassis/test_utils"
	"github.com/veganbase/backend/services/user-service/db"
	"github.com/veganbase/backend/services/user-service/model"
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

func TestUserRetrieval(t *testing.T) {
	RunWithSchema(t, func(pg *db.PGClient, t *testing.T) {
		loadDefaultFixture(pg, t)

		var tests = []struct {
			id    string
			err   error
			email string
		}{
			{"usr_TESTUSER1", nil, "test@example.com"},
			{"usr_UNKNOWN", db.ErrUserNotFound, ""},
		}
		for _, test := range tests {
			user, err := pg.UserByID(test.id)
			assert.Equal(t, err, test.err)
			if err != nil {
				continue
			}
			if assert.NotNil(t, user) {
				assert.Equal(t, user.Email, test.email, "wrong user!")
			}
		}
	})
}

func TestLoginUser(t *testing.T) {
	RunWithSchema(t, func(pg *db.PGClient, t *testing.T) {
		loadDefaultFixture(pg, t)

		var tests = []struct {
			email   string
			userID  string
			newUser bool
		}{
			{"test@example.com", "usr_TESTUSER1", false},
			{"newuser@example.com", "", true},
		}
		for _, test := range tests {
			user, newUser, err := pg.LoginUser(test.email, func() string {
				return "https://www.avatar.net/avatar.png"
			})
			assert.Nil(t, err)
			assert.Equal(t, newUser, test.newUser, "new user mismatch")
			if test.email != "" {
				assert.Equal(t, user.Email, test.email, "email mismatch")
			}
		}
	})
}

func TestUpdateUser(t *testing.T) {
	RunWithSchema(t, func(pg *db.PGClient, t *testing.T) {
		loadDefaultFixture(pg, t)

		// Check that updating name, etc. work.
		user, err := pg.UserByID("usr_TESTUSER1")
		assert.Nil(t, err)
		newName := "Update test"
		user.Name = &newName
		assert.Nil(t, pg.UpdateUser(user))

		// Check that updating with bad ID fails.
		user.ID = "usr_UNKNOWN"
		newName = "Update test 2"
		user.Name = &newName
		assert.Equal(t, pg.UpdateUser(user), db.ErrUserNotFound)

		// Check that updating email fails.
		user, err = pg.UserByID("usr_TESTUSER1")
		assert.Nil(t, err)
		user.Email = "new@somewhere.com"
		assert.Equal(t, pg.UpdateUser(user), db.ErrReadOnlyField,
			"update of read-only field!")
	})
}

func TestDeleteUser(t *testing.T) {
	RunWithSchema(t, func(pg *db.PGClient, t *testing.T) {
		loadDefaultFixture(pg, t)

		var tests = []struct {
			id  string
			err error
		}{
			{"usr_TESTUSER1", nil},
			{"usr_UNKNOWN", db.ErrUserNotFound},
		}
		for _, test := range tests {
			err := pg.DeleteUser(test.id)
			assert.Equal(t, err, test.err)
			if err != nil {
				continue
			}
			_, err = pg.UserByID(test.id)
			assert.Equal(t, err, db.ErrUserNotFound, "user not deleted")
		}
	})
}

func TestAPIKeys(t *testing.T) {
	RunWithSchema(t, func(pg *db.PGClient, t *testing.T) {
		loadDefaultFixture(pg, t)
		apikey := "testApiKey1234567"
		apikey2 := "testApiKey123456789101112"
		secret := "maskedAPiSecret1234567"
		user, err := pg.UserByID("usr_TESTUSER1")
		assert.Nil(t, err)
		assert.Nil(t, user.APIKey, "already has API key")

		err = pg.SaveHashedAPIKey(apikey, secret, "usr_TESTUSER1")
		assert.Nil(t, err)
		user, err = pg.UserByID("usr_TESTUSER1")
		assert.Nil(t, err)
		assert.Equal(t, *user.APIKey, apikey, "API key mismatch")

		err = pg.SaveHashedAPIKey(apikey2, secret, "usr_TESTUSER1")
		assert.Nil(t, err)
		user, err = pg.UserByID("usr_TESTUSER1")
		assert.Nil(t, err)
		assert.Equal(t, *user.APIKey, apikey2, "API key mismatch")

		err = pg.DeleteAPIKey("usr_TESTUSER1")
		assert.Nil(t, err)
		user, err = pg.UserByID("usr_TESTUSER1")
		assert.Nil(t, err)
		assert.Nil(t, user.APIKey, "API key not deleted")
	})
}

func TestListUsers(t *testing.T) {
	RunWithSchema(t, func(pg *db.PGClient, t *testing.T) {
		loadDefaultFixture(pg, t)

		// Get all users: test for count and ordering.
		users, err := pg.Users("", 1, 30)
		assert.Nil(t, err)
		assert.Len(t, users, 10)
		assertOrdered(t, users)

		// Pagination: test for count and ordering.
		users, err = pg.Users("", 1, 4)
		assert.Nil(t, err)
		assert.Len(t, users, 4)
		assertOrdered(t, users)
		users, err = pg.Users("", 2, 4)
		assert.Nil(t, err)
		assert.Len(t, users, 4)
		assertOrdered(t, users)

		// Search.
		users, err = pg.Users("Target", 1, 30)
		assert.Nil(t, err)
		assert.Len(t, users, 2)
		assertOrdered(t, users)
	})
}

func assertOrdered(t *testing.T, users []model.User) {
	for i, u := range users {
		if i > 0 {
			assert.True(t, u.LastLogin.Before(users[i-1].LastLogin))
		}
	}
}
