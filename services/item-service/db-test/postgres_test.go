// The following environment variables will be used to get a Postgres
// connection string: VB_TEST_DB.
//
package db

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/rs/zerolog/log"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/stretchr/testify/assert"
	"github.com/veganbase/backend/chassis/test_utils"
	"github.com/veganbase/backend/services/item-service/db"
	"github.com/veganbase/backend/services/item-service/model"
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

// TODO: FILL THESE IN

func TestItemRetrieval(t *testing.T) {
	RunWithSchema(t, func(pg *db.PGClient, t *testing.T) {
		loadDefaultFixture(pg, t)

		var tests = []struct {
			id   string
			slug string
			err  error
			name string
		}{
			{"htl_EEOBepsfpJKH0EPW", "the-vegan-lodge", nil, "The Vegan Lodge"},
			{"pfd_VmV3Pry2y0aavzaq", "beyond-burger", nil, "Beyond Burger"},
			{"pfd_VmV3Pry2y0aavzdx", "bad-slug", db.ErrItemNotFound, ""},
		}

		for _, test := range tests {
			item, err := pg.ItemByID(test.id)
			assert.Equal(t, err, test.err)
			if err != nil {
				continue
			}
			if assert.NotNil(t, item) {
				assert.Equal(t, item.Slug, test.slug, "wrong item slug!")
				assert.Equal(t, item.Name, test.name, "wrong item name!")
			}
		}

		for _, test := range tests {
			item, err := pg.ItemBySlug(test.slug)
			assert.Equal(t, err, test.err)
			if err != nil {
				continue
			}
			if assert.NotNil(t, item) {
				assert.Equal(t, item.ID, test.id, "wrong item ID!")
				assert.Equal(t, item.Name, test.name, "wrong item name!")
			}
		}

		for _, test := range tests {
			item, err := pg.ItemByIDOrSlug(test.id)
			assert.Equal(t, err, test.err)
			if err != nil {
				continue
			}
			if assert.NotNil(t, item) {
				assert.Equal(t, item.Slug, test.slug, "wrong item slug!")
				assert.Equal(t, item.Name, test.name, "wrong item name!")
			}
		}

		for _, test := range tests {
			item, err := pg.ItemByIDOrSlug(test.slug)
			assert.Equal(t, err, test.err)
			if err != nil {
				continue
			}
			if assert.NotNil(t, item) {
				assert.Equal(t, item.ID, test.id, "wrong item ID!")
				assert.Equal(t, item.Name, test.name, "wrong item name!")
			}
		}
	})
}

func TestListItems(t *testing.T) {
	RunWithSchema(t, func(pg *db.PGClient, t *testing.T) {
		loadDefaultFixture(pg, t)

		it := []model.ItemType{model.HotelItem}
		params := db.SearchParams{
			ItemTypes: &it,
		}
		items, _ , err := pg.SummaryItems(&params, nil, nil, 1, 1000)
		assert.Nil(t, err)
		assert.Len(t, items, 24)

		it = []model.ItemType{model.RestaurantItem}
		tag := "pizza"
		params = db.SearchParams{
			ItemTypes: &it,
			Tag:      &tag,
		}
		items, _, err = pg.SummaryItems(&params, nil, nil, 1, 1000)
		assert.Nil(t, err)
		assert.Len(t, items, 3)
	})
}

func TestCreateItem(t *testing.T) {
	RunWithSchema(t, func(pg *db.PGClient, t *testing.T) {
		loadDefaultFixture(pg, t)

	})
}

func TestUpdateItem(t *testing.T) {
	RunWithSchema(t, func(pg *db.PGClient, t *testing.T) {
		loadDefaultFixture(pg, t)

		item, err := pg.ItemByID("htl_EEOBepsfpJKH0EPW")
		assert.Nil(t, err)
		item.Slug = "bad-slug"
		err = pg.UpdateItem(item, []string{})
		assert.Equal(t, err, db.ErrReadOnlyField)

		item, err = pg.ItemByID("htl_EEOBepsfpJKH0EPW")
		assert.Nil(t, err)
		slug := item.Slug
		taglen := len(item.Tags)

		item.Tags = append(item.Tags, "test-tag")
		item.Name = "Check Name Change"
		err = pg.UpdateItem(item, []string{})
		assert.Nil(t, err)

		item, err = pg.ItemByID("htl_EEOBepsfpJKH0EPW")
		assert.Nil(t, err)
		assert.Equal(t, len(item.Tags), taglen+1)
		assert.NotEqual(t, item.Slug, slug)
	})
}

func TestDeleteItem(t *testing.T) {
	RunWithSchema(t, func(pg *db.PGClient, t *testing.T) {
		loadDefaultFixture(pg, t)

		pics, err := pg.DeleteItem("htl_EEOBepsfpJKH0EPW", []string{})
		assert.Nil(t, err)
		assert.Len(t, pics, 17)

		_, err = pg.ItemByID("htl_EEOBepsfpJKH0EPW")
		assert.Equal(t, err, db.ErrItemNotFound)
	})
}

func TestUserTags(t *testing.T) {
	RunWithSchema(t, func(pg *db.PGClient, t *testing.T) {
		loadDefaultFixture(pg, t)

		userID := "usr_AGsNFyCDfjshffNJ"

		params := db.SearchParams{Owner: []string{userID}}
		items, _, err := pg.SummaryItems(&params, nil, nil, 1, 1000)
		assert.Nil(t, err)
		tagMap1 := map[string]bool{}
		for _, it := range items {
			for _, tag := range it.Tags {
				tagMap1[tag] = true
			}
		}

		tags, err := pg.TagsForUser(userID)
		assert.Nil(t, err)
		tagMap2 := map[string]bool{}
		for _, tag := range tags {
			tagMap2[tag] = true
		}

		assert.Equal(t, len(tagMap1), len(tagMap2))
		for tag := range tagMap1 {
			assert.Equal(t, tagMap1[tag], tagMap2[tag])
		}
	})
}

func TestCreateLink(t *testing.T) {

}

func TestDeleteLink(t *testing.T) {

}
