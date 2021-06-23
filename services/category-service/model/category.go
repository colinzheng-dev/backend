package model

import (
	"time"

	"github.com/jmoiron/sqlx/types"
)

// Category models a mapping between category keys and values
// conforming to a JSON schema.
type Category struct {
	// The textual name of the category.
	ID string `db:"id"`

	// A human-readable textual label for the category.
	Label string `db:"label"`

	// A flag saying whether the entries in a category can be expanded
	// by normal users. If it is false, only administrators may add new
	// values to the category.
	Extensible bool `db:"extensible"`

	// The JSON schema for the category.
	Schema types.JSONText `db:"schema"`

	// The creation time for the category.
	CreatedAt time.Time `db:"created_at"`
}
