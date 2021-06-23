// The following environment variables will be used to get a Postgres
// connection string: VB_TEST_DB.
//
package db

import (
	"context"
	"io/ioutil"
	"os"
	"sort"
	"testing"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/stretchr/testify/assert"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/chassis/test_utils"
	"github.com/veganbase/backend/services/blob-service/db"
	"github.com/veganbase/backend/services/blob-service/model"
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

func TestSingleBlobRetrieval(t *testing.T) {
	RunWithSchema(t, func(pg *db.PGClient, t *testing.T) {
		loadDefaultFixture(pg, t)

		var tests = []struct {
			id     string
			err    error
			user   string
			tags   int
			assocs int
		}{
			{"blob0001", nil, "usr_TESTUSER1", 2, 2},
			{"blob0004", nil, "usr_TESTUSER3", 0, 0},
			{"blob000X", db.ErrBlobNotFound, "", 0, 0},
		}
		for _, test := range tests {
			blob, err := pg.BlobByID(test.id)
			assert.Equal(t, err, test.err)
			if err != nil {
				continue
			}
			assert.NotNil(t, blob)
			assert.NotNil(t, blob.Owner)
			assert.Equal(t, *blob.Owner, test.user)
			assert.Equal(t, len(blob.Tags), test.tags)
			assert.Equal(t, len(blob.AssociatedItems), test.assocs)
		}
	})
}

func TestMultiBlobRetrieval(t *testing.T) {
	RunWithSchema(t, func(pg *db.PGClient, t *testing.T) {
		loadDefaultFixture(pg, t)

		var tests = []struct {
			user    string
			tags    []string
			page    uint
			perPage uint
			ids     []string
		}{
			{"usr_TESTUSER1", nil, 0, 0, []string{"blob0001", "blob0005", "blob0006", "blob0009", "blob0010"}},
			{"usr_TESTUSER1", nil, 0, 2, []string{"blob0001", "blob0005"}},
			{"usr_TESTUSER1", nil, 2, 2, []string{"blob0006", "blob0009"}},
			{"usr_TESTUSER1", []string{"beach"}, 0, 0, []string{"blob0001", "blob0006", "blob0010"}},
			{"usr_TESTUSER2", []string{"lunch", "shells"}, 0, 0, []string{"blob0002", "blob0003", "blob0008"}},
		}
		for _, test := range tests {
			blobs, err := pg.BlobsByUser(test.user, test.tags, test.page, test.perPage)
			assert.Nil(t, err)
			assert.Equal(t, len(blobs), len(test.ids), "wrong number of blobs")
			for i, blob := range blobs {
				assert.Equal(t, blob.ID, test.ids[i], "wrong IDs!")
			}
		}
	})
}

func TestCreateBlob(t *testing.T) {
	RunWithSchema(t, func(pg *db.PGClient, t *testing.T) {
		loadDefaultFixture(pg, t)

		user := "usr_TESTUSER1"
		blob := model.Blob{
			Format: "image/jpeg",
			Size:   234556,
			Owner:  &user,
			Tags:   chassis.Tags([]string{"flower", "bees"}),
		}
		assert.Nil(t, pg.CreateBlob(&blob))
	})
}

func TestBlobTagsModification(t *testing.T) {
	RunWithSchema(t, func(pg *db.PGClient, t *testing.T) {
		loadDefaultFixture(pg, t)

		var tests = []struct {
			id   string
			err  error
			tags []string
		}{
			{"blob0004", nil, []string{"test1", "test2"}},
			{"blob0001", nil, []string{}},
			{"blob000X", db.ErrBlobNotFound, nil},
		}
		for _, test := range tests {
			err := pg.SetBlobTags(test.id, test.tags)
			assert.Equal(t, err, test.err)
			if err != nil {
				continue
			}
			blob, err := pg.BlobByID(test.id)
			assert.Nil(t, err)
			assert.Equal(t, len(blob.Tags), len(test.tags), "wrong number of tags")
			sort.Strings(blob.Tags)
			sort.Strings(test.tags)
			for i, tag := range blob.Tags {
				assert.Equal(t, tag, test.tags[i], "wrong tags")
			}
		}
	})
}

func TestClearBlobOwner(t *testing.T) {
	RunWithSchema(t, func(pg *db.PGClient, t *testing.T) {
		loadDefaultFixture(pg, t)

		var tests = []struct {
			id     string
			err    error
			result bool
		}{
			{"blob0001", nil, true},
			{"blob0003", nil, false},
			{"blob000X", db.ErrBlobNotFound, false},
		}
		for _, test := range tests {
			result, err := pg.ClearBlobOwner(test.id)
			assert.Equal(t, err, test.err)
			if err != nil {
				continue
			}
			assert.Equal(t, result, test.result, "wrong result")
			blob, err := pg.BlobByID(test.id)
			assert.Nil(t, err)
			assert.Nil(t, blob.Owner, "didn't clear owner")
		}
	})
}

func TestTagsForUser(t *testing.T) {
	RunWithSchema(t, func(pg *db.PGClient, t *testing.T) {
		loadDefaultFixture(pg, t)

		var tests = []struct {
			userID string
			tags   []string
		}{
			{"usr_TESTUSER1", []string{"beach", "holiday"}},
			{"usr_TESTUSER2", []string{"lunch", "cake", "cordon-bleu", "seaside", "shells"}},
			{"usr_TESTUSER4", []string{}},
			{"usr_TESTUSERX", []string{}},
		}
		for _, test := range tests {
			tags, err := pg.TagsForUser(test.userID)
			assert.Nil(t, err)
			assert.Equal(t, len(tags), len(test.tags), "wrong number of tags")
			sort.Strings(tags)
			sort.Strings(test.tags)
			for i, tag := range tags {
				assert.Equal(t, tag, test.tags[i], "wrong tags")
			}
		}
	})
}

func TestDeleteBlob(t *testing.T) {
	RunWithSchema(t, func(pg *db.PGClient, t *testing.T) {
		loadDefaultFixture(pg, t)

		var tests = []struct {
			id  string
			err error
		}{
			{"blob0001", nil},
			{"blob0006", nil},
			{"blob000X", db.ErrBlobNotFound},
		}
		for _, test := range tests {
			err := pg.DeleteBlob(test.id)
			assert.Equal(t, err, test.err)
			if err != nil {
				continue
			}
			count := 0
			err = pg.DB.Get(&count,
				`SELECT COUNT(*) FROM blob_items WHERE blob_id = $1`, test.id)
			assert.Nil(t, err)
			assert.Equal(t, count, 0, "didn't delete associations")
		}
	})
}

func TestBlobItemAssociations(t *testing.T) {
	RunWithSchema(t, func(pg *db.PGClient, t *testing.T) {
		loadDefaultFixture(pg, t)

		var tests = []struct {
			add    bool
			blobID string
			itemID string
			count  int
			inUse  bool
		}{
			{true, "blob0004", "item0008", 1, true},
			{true, "blob0001", "item0008", 3, true},
			{false, "blob0013", "item0025", 2, true},
			{false, "blob0013", "item0026", 1, true},
			{false, "blob0013", "item0023", 0, false},
			{false, "blob0014", "item0080", 0, false},
			{true, "blob0001", "item0001", 3, true},
		}
		for _, test := range tests {
			var err error
			deleted := []db.DeletedBlob{}
			blobIDs := []string{test.blobID}
			if test.add {
				err = pg.AddBlobsToItem(test.itemID, blobIDs)
			} else {
				deleted, err = pg.RemoveBlobsFromItem(test.itemID, blobIDs)
			}
			assert.Nil(t, err)
			if err != nil {
				continue
			}
			if test.inUse {
				assert.Equal(t, len(deleted), 0, "in use mismatch")
			} else {
				assert.Equal(t, len(deleted), 1, "in use mismatch")
			}
			count := 0
			err = pg.DB.Get(&count,
				`SELECT COUNT(*) FROM blob_items WHERE blob_id = $1`, test.blobID)
			assert.Nil(t, err)
			assert.Equal(t, count, test.count, "mismatching associations")
		}
	})
}

func TestDeleteByItem(t *testing.T) {
	RunWithSchema(t, func(pg *db.PGClient, t *testing.T) {
		loadDefaultFixture(pg, t)

		var tests = []struct {
			itemID string
			err    error
			ids    []string
		}{
			{"item0001", nil, nil},
			{"item0080", nil, []string{"blob0014", "blob0015"}},
		}
		for _, test := range tests {
			deleted, err := pg.RemoveBlobsFromItem(test.itemID, []string{})
			assert.Equal(t, err, test.err)
			if err != nil {
				continue
			}
			ids := []string{}
			for _, d := range deleted {
				ids = append(ids, d.ID)
			}
			assert.Equal(t, len(ids), len(test.ids), "ID count mismatch")
			sort.Strings(ids)
			sort.Strings(test.ids)
			for i, id := range ids {
				assert.Equal(t, id, test.ids[i], "ID mismatch")
			}
		}
	})
}
