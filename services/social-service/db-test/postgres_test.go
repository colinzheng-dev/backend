package db_test

import (
	"context"
	"github.com/rs/zerolog/log"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/veganbase/backend/chassis/test_utils"
	"github.com/veganbase/backend/services/social-service/db"
	"io/ioutil"
	"os"
	"testing"
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
	f, err := os.Open("fixtures.sql")
	assert.Nil(t, err)
	defer f.Close()
	fixture, err := ioutil.ReadAll(f)
	assert.Nil(t, err)
	tx := pg.DB.MustBegin()
	test_utils.MultiExec(tx, string(fixture))
	tx.Commit()
}

func TestCreateSubscription(t *testing.T) {
	RunWithSchema(t, func(pg *db.PGClient, t *testing.T) {
		userID := "usr_AGsNFyCDfjshffNJ"
		subscriptionID := "usr_tPOMyx1UV2wdgMT8"

		err := pg.CreateUserSubscription(userID, subscriptionID)

		require.Nil(t, err)

		subs, err := pg.ListUserSubscriptions(userID)

		assert.NotEmpty(t, subs)
		require.Len(t, subs, 1)

		assert.Equal(t, subscriptionID, subs[0])
	})
}

func TestListSubscriptions(t *testing.T) {
	RunWithSchema(t, func(pg *db.PGClient, t *testing.T) {
		loadDefaultFixture(pg, t)

		knownUserId := "usr_AGsNFyCDfjshffNJ"
		knownSubscriptions := []string{
			"usr_tPOMyx1UV2wdgMT8",
			"usr_mO1Dld1gn6s9Itdv",
			"usr_KJDDyBHthY4Vh7y9",
		}

		subs, err := pg.ListUserSubscriptions(knownUserId)

		assert.Nil(t, err)
		assert.Len(t, subs, len(knownSubscriptions))
		assert.ElementsMatch(t, subs, knownSubscriptions)
	})
}

func TestDeleteSubscription(t *testing.T)  {
	RunWithSchema(t, func(pg *db.PGClient, t *testing.T) {
		loadDefaultFixture(pg, t)

		knownUserId := "usr_AGsNFyCDfjshffNJ"
		subscriptionToDelete := "usr_tPOMyx1UV2wdgMT8"
		remainingSubscriptions := []string{
			"usr_mO1Dld1gn6s9Itdv",
			"usr_KJDDyBHthY4Vh7y9",
		}

		err := pg.DeleteUserSubscription(knownUserId, subscriptionToDelete)

		assert.Nil(t, err)

		subs, err := pg.ListUserSubscriptions(knownUserId)

		require.Len(t, subs, len(remainingSubscriptions))
		require.ElementsMatch(t, subs, remainingSubscriptions)
	})
}


func TestListFollowers(t *testing.T) {
	RunWithSchema(t, func(pg *db.PGClient, t *testing.T) {
		loadDefaultFixture(pg, t)

		subscriptionID := "usr_KJDDyBHthY4Vh7y9"
		userIDs := []string{
			"usr_Rjzyq0UFFrYiYcHq",
			"usr_xi1L3id1SR5OOZee",
			"usr_AGsNFyCDfjshffNJ",
		}

		followers, err := pg.ListFollowers(subscriptionID)

		assert.Nil(t, err)
		assert.Len(t, followers, len(userIDs))
		assert.ElementsMatch(t, followers, userIDs)
	})
}
