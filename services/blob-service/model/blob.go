package model

import (
	"time"

	"github.com/veganbase/backend/chassis"
)

// Blob represents binary data uploaded by a user (usually an image).
type Blob struct {
	// Unique ID of the blob.
	ID string `json:"id" db:"id"`

	// In the database, this stores the Google storage URL of the blob.
	// In messages returned by the API, this is the access URL for the
	// blob that goes via the image reverse proxy. We don't expose the
	// Google Storage URL externally.
	URI string `json:"uri" db:"uri"`

	// The file extension for the type of the blob.
	Format string `json:"format" db:"format"`

	// The size of the blob in bytes.
	Size int `json:"size" db:"size"`

	// The user ID of the owner of the blob.
	Owner *string `json:"owner" db:"owner"`

	// Metadata tags applied to the blob.
	Tags chassis.Tags `json:"tags" db:"tags"`

	// The creation date of the blob.
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// IDs of items this blob is associated with.
	AssociatedItems []string `json:"associated_items" db:""`
}
