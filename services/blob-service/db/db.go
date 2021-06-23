package db

import "github.com/veganbase/backend/services/blob-service/model"

// DeletedBlob carries information about blobs deleted from the
// database that is needed to delete the associated entries from the
// blob storage.
type DeletedBlob struct {
	ID     string
	Format string
}

// DB describes the database operations used by the blob service.
type DB interface {
	// BlobByID returns the model for a given blob ID.
	BlobByID(id string) (*model.Blob, error)

	// BlobsByUser get a list of blobs in reverse creation date order
	// for a given user, optionally filtered by tag and paginated.
	BlobsByUser(userID string, tags []string, page, perPage uint) ([]model.Blob, error)

	// NewBlobID creates a new ID for a blob. This is needed because the
	// blob ID is created outside of the CreateBlob function to allow
	// for uploading blob data to storage before creating a database
	// record.
	NewBlobID() string

	// CreateBlob creates a new blob.
	CreateBlob(blob *model.Blob) error

	// SetBlobTags sets the tag list for a blob.
	SetBlobTags(id string, tags []string) error

	// ClearBlobOwner deletes the given blob from the image gallery of its
	// owning user, returning a boolean flag indicating whether the blob
	// is still in use in association with items.
	ClearBlobOwner(id string) (bool, error)

	// DeleteBlob performs the actual deletion of a blob and any item
	// associations.
	DeleteBlob(id string) error

	// TagsForUser returns all the tags used in blobs owned by a user.
	TagsForUser(userID string) ([]string, error)

	// AddBlobsToItem adds associations between a list of blobs and an
	// item.
	AddBlobsToItem(itemID string, blobIDs []string) error

	// RemoveBlobsFromItem removes the associations between an item and
	// a list of blobs, returning information about unused blobs that
	// were deleted as a result of this action.
	RemoveBlobsFromItem(id string, itemIDs []string) ([]DeletedBlob, error)

	// SaveEvent saves an event to the database.
	SaveEvent(topic string, eventData interface{}, inTx func() error) error
}

//go:generate go-bindata -pkg db -o migrations.go migrations/...
