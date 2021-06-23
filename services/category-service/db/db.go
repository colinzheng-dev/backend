package db

import (
	"errors"

	"github.com/jmoiron/sqlx/types"
	"github.com/veganbase/backend/services/category-service/model"
)

// ErrCategoryNotFound is the error returned when an attempt is made
// to access or manipulate a category with an unknown ID.
var ErrCategoryNotFound = errors.New("category name not found")

// ErrCategoryEntryNotFound is the error returned when an attempt is
// made to access or manipulate an unknown category entry.
var ErrCategoryEntryNotFound = errors.New("category entry not found")

// ErrCategoryLabelNotUnique is the error returned when an attempt is
// made to add an entry to a category for a label that already exists.
var ErrCategoryLabelNotUnique = errors.New("category label is not unique")

// ErrSchemaMismatch is the error returned when data submitted for a
// category value does not match the category's schema.
var ErrSchemaMismatch = errors.New("data does not match category schema")

// EntryInfo contains information about a category entry: whether or
// not the entry is fixed, and the user ID of the user that created
// the entry.
type EntryInfo struct {
	Fixed   bool   `db:"fixed"`
	Creator string `db:"creator"`
}

// DB describes the database operations used by the user service.
type DB interface {
	// Categories gets the list of categories and their JSON schemas.
	Categories() (map[string]*model.Category, error)

	// CategoryByName looks up a category by name.
	CategoryByName(name string) (*model.Category, error)

	// CategoryEntries gets all the entries for a category, returning a
	// map from category labels to category values conforming to the
	// category's JSON schema.
	CategoryEntries(category string, fixed *bool) (map[string]interface{}, error)

	// EntryInfo returns information about an entry in a category,
	// specifically whether or not the entry exists (non-nil or nil
	// EntryInfo return) and the fixation status and creating user ID
	// for the entry.
	EntryInfo(category string, label string) (*EntryInfo, error)

	// AddCategoryEntry adds an entry for a category.
	AddCategoryEntry(category string, label string, value types.JSONText, creator string) error

	// FixCategoryEntry updates the fixation state for an entry for a
	// category.
	FixCategoryEntry(category string, label string, fixed bool) error

	// UpdateCategoryEntry updates an entry for a category.
	UpdateCategoryEntry(category string, label string, value types.JSONText) error

	// SaveEvent saves an event to the database.
	SaveEvent(label string, eventData interface{}, inTx func() error) error
}

//go:generate go-bindata -pkg db -o migrations.go migrations/...
