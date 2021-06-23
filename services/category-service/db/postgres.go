package db

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/types"
	gojs "github.com/xeipuuv/gojsonschema"

	// Import Postgres DB driver.
	_ "github.com/lib/pq"

	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/category-service/model"
)

// PGClient is a wrapper for the category database connection.
type PGClient struct {
	DB *sqlx.DB
}

// NewPGClient creates a new category database connection.
func NewPGClient(ctx context.Context, dbURL string) (*PGClient, error) {
	db, err := chassis.DBConnect(ctx, "category", dbURL, Asset, AssetDir)
	if err != nil {
		return nil, err
	}
	return &PGClient{db}, nil
}

// Categories retrieves the category list.
func (pg *PGClient) Categories() (map[string]*model.Category, error) {
	categories := []model.Category{}
	err := pg.DB.Select(&categories,
		`SELECT id, label, extensible, schema, created_at FROM categories`)
	if err != nil {
		return nil, err
	}
	categorymap := map[string]*model.Category{}
	for _, s := range categories {
		s := s
		categorymap[s.ID] = &s
	}
	return categorymap, nil
}

// CategoryByName looks up a category by name.
func (pg *PGClient) CategoryByName(name string) (*model.Category, error) {
	category := model.Category{}
	err := pg.DB.Get(&category, categoryByName, name)
	fmt.Println("--->", err)
	if err != nil {
		return nil, ErrCategoryNotFound
	}
	return &category, nil
}

const categoryByName = `
SELECT id, label, extensible, schema, created_at
  FROM categories WHERE id = $1`

// CategoryEntries gets all the entries for a category, returning a
// map from category labels to category values conforming to the
// category's JSON schema.
func (pg *PGClient) CategoryEntries(category string, fixed *bool) (map[string]interface{}, error) {
	json := types.JSONText{}
	var err error
	if fixed == nil {
		err = pg.DB.Get(&json, categoryEntries, category)
	} else {
		err = pg.DB.Get(&json, categoryEntriesWithFixed, category, *fixed)
	}
	if err != nil {
		return nil, err
	}

	entries := map[string]interface{}{}
	if err = json.Unmarshal(&entries); err != nil {
		return nil, err
	}

	if len(entries) == 0 {
		return nil, ErrCategoryNotFound
	}

	return entries, nil
}

const categoryEntries = `
SELECT json_object_agg(label, value)
  FROM category_entries
 WHERE category = $1 AND lang = 'en'`

const categoryEntriesWithFixed = `
SELECT json_object_agg(label, value)
  FROM category_entries
 WHERE category = $1 AND lang = 'en' AND fixed = $2`

// EntryInfo returns information about an entry in a category,
// specifically whether or not the entry exists (non-nil or nil
// EntryInfo return) and the fixation status and creating user ID
// for the entry.
func (pg *PGClient) EntryInfo(category string, label string) (*EntryInfo, error) {
	info := []EntryInfo{}
	err := pg.DB.Select(&info, entryInfo, category, label)
	if err != nil {
		return nil, err
	}
	if len(info) == 0 {
		return nil, nil
	}
	return &info[0], nil
}

const entryInfo = `
SELECT fixed, creator FROM category_entries
 WHERE category = $1 AND lang = 'en' AND label = $2`

func checkEntryEdit(tx *sqlx.Tx, category string, label string,
	value types.JSONText) error {
	// Look up category.
	cat := model.Category{}
	err := tx.Get(&cat, categoryByName, category)
	if err != nil {
		return ErrCategoryNotFound
	}

	// Validate request body against category schema.
	loader := gojs.NewSchemaLoader()
	schema, err := loader.Compile(gojs.NewStringLoader(cat.Schema.String()))
	if err != nil {
		return err
	}
	res, err := schema.Validate(gojs.NewBytesLoader(value))
	if err != nil {
		return err
	}
	if !res.Valid() {
		return ErrSchemaMismatch
	}
	return nil
}

// AddCategoryEntry adds an entry for a category.
func (pg *PGClient) AddCategoryEntry(category string, label string,
	value types.JSONText, creator string) error {
	tx, err := pg.DB.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			tx.Rollback()
		}
	}()

	err = checkEntryEdit(tx, category, label, value)
	if err != nil {
		return err
	}

	// Add entry to category.
	addres, err := tx.Exec(addCategoryEntry, category, label, creator, value)
	if err != nil {
		return err
	}
	rows, err := addres.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ErrCategoryLabelNotUnique
	}
	return nil
}

const addCategoryEntry = `
INSERT INTO category_entries (category, label, creator, value)
VALUES ($1, $2, $3, $4)
ON CONFLICT DO NOTHING`

// UpdateCategoryEntry updates an entry for a category.
func (pg *PGClient) UpdateCategoryEntry(category string, label string,
	value types.JSONText) error {
	tx, err := pg.DB.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			tx.Rollback()
		}
	}()

	err = checkEntryEdit(tx, category, label, value)
	if err != nil {
		return err
	}

	// Add entry to category.
	_, err = tx.Exec(updateCategoryEntry, category, label, value)
	if err != nil {
		return err
	}
	return nil
}

const updateCategoryEntry = `
UPDATE category_entries SET value = $3
 WHERE category = $1 AND label = $2`

// FixCategoryEntry updates the fixation state for an entry for a
// category.
func (pg *PGClient) FixCategoryEntry(category string, label string, fixed bool) error {
	res, err := pg.DB.Exec(fixCategoryEntry, category, label, fixed)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ErrCategoryEntryNotFound
	}
	return nil
}

const fixCategoryEntry = `
UPDATE category_entries SET fixed = $3
 WHERE category = $1 AND label = $2`

// SaveEvent saves an event to the database.
func (pg *PGClient) SaveEvent(label string, eventData interface{}, inTx func() error) error {
	tx, err := pg.DB.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			tx.Rollback()
		}
	}()
	err = chassis.SaveEvent(tx, label, eventData, inTx)
	return err
}
