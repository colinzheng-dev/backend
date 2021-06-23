package db

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/blob-service/model"
)

// ErrBlobNotFound is the error returned when an attempt is made to
// access or manipulate a blob with an unknown ID.
var ErrBlobNotFound = errors.New("blob ID not found")

// ErrAssocNotFound is the error returned when an attempt is made to
// access or manipulate an unknown blob/item association.
var ErrAssocNotFound = errors.New("blob/item association not found")

// PGClient is a wrapper for the user database connection.
type PGClient struct {
	DB *sqlx.DB
}

// NewPGClient creates a new user database connection.
func NewPGClient(ctx context.Context, dbURL string) (*PGClient, error) {
	db, err := chassis.DBConnect(ctx, "blob", dbURL, Asset, AssetDir)
	if err != nil {
		return nil, err
	}
	return &PGClient{db}, nil
}

// BlobByID looks up a blob by its blog ID.
func (pg *PGClient) BlobByID(id string) (*model.Blob, error) {
	tx, err := pg.DB.Beginx()
	if err != nil {
		return nil, err
	}
	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			tx.Rollback()
		}
	}()

	blob := &model.Blob{}
	err = tx.Get(blob, blobByID, id)
	if err == sql.ErrNoRows {
		return nil, ErrBlobNotFound
	}
	if err != nil {
		return nil, err
	}
	return addItemAssociations(tx, blob)
}

const blobByID = `
SELECT id, format, size, owner, tags, created_at
  FROM blobs WHERE id = $1`

// TODO: MAKE THIS NICER. THIS CAN BE DONE WITH A SINGLE QUERY, BUT
// NEEDS A LITTLE WORK TO DEAL WITH THE FACT THAT THE BLOB
// ASSOCIATEDITEMS FIELD DOESN'T HAVE A DB TAG BECAUSE WE DON'T KEEP
// IT IN THE BLOBS TABLE. THERE HAS TO BE A GOOD WAY TO DO THIS, BUT I
// DON'T HAVE TIME TO FIGURE IT OUT RIGHT NOW.
//
// const blobByID = `
// SELECT b.id, b.uri, b.format, b.size, b.owner, b.tags, b.created_at,
//        array_agg(i.item_id) AS associated_items
//   FROM blobs b JOIN blob_items i ON b.id = i.blob_id
//  WHERE b.id = $1 GROUP BY b.id`

// BlobsByUser gets a list of blobs in reverse creation date order for
// a given user, optionally filtered by tag and paginated.
func (pg *PGClient) BlobsByUser(userID string,
	tags []string, page, perPage uint) ([]model.Blob, error) {
	tx, err := pg.DB.Beginx()
	if err != nil {
		return nil, err
	}
	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			tx.Rollback()
		}
	}()

	results := []model.Blob{}
	if tags != nil && len(tags) > 0 {
		err = tx.Select(&results,
			blobsByUserWithTags+chassis.Paginate(page, perPage), userID, pq.Array(tags))
	} else {
		err = tx.Select(&results,
			blobsByUser+chassis.Paginate(page, perPage), userID)
	}
	if err != nil {
		return nil, err
	}

	for _, blob := range results {
		_, err = addItemAssociations(tx, &blob)
		if err != nil {
			return nil, err
		}
	}
	return results, nil
}

const blobsByUser = `
SELECT id, format, size, owner, tags, created_at
  FROM blobs WHERE owner = $1
 ORDER BY created_at DESC`

const blobsByUserWithTags = `
SELECT id, format, size, owner, tags, created_at
  FROM blobs WHERE owner = $1 AND tags ?| $2
 ORDER BY created_at DESC`

// Add associated items to blob information.
func addItemAssociations(tx *sqlx.Tx, blob *model.Blob) (*model.Blob, error) {
	itemIDs := []string{}
	err := tx.Select(&itemIDs, itemAssociations, blob.ID)
	if err != nil {
		return nil, err
	}
	blob.AssociatedItems = itemIDs
	return blob, nil
}

const itemAssociations = `
SELECT item_id FROM blob_items WHERE blob_id = $1`

// NewBlobID creates a new ID for a blob. This is needed because the
// blob ID is created outside of the CreateBlob function to allow for
// uploading blob data to storage before creating a database record.
func (pg *PGClient) NewBlobID() string {
	return chassis.NewBareID(16)
}

// CreateBlob creates a new blob.
func (pg *PGClient) CreateBlob(blob *model.Blob) error {
	rows, err := pg.DB.NamedQuery(createBlob, blob)
	if err != nil {
		return err
	}
	defer rows.Close()
	if !rows.Next() {
		return sql.ErrNoRows
	}
	err = rows.Scan(&blob.CreatedAt)
	if err != nil {
		return err
	}
	return nil
}

func extensionForMIMEType(mime string) string {
	switch mime {
	case "image/jpeg":
		return "jpg"
	case "image/png":
		return "png"
	default:
		return "dat"
	}
}

const createBlob = `
INSERT INTO blobs (id, uri, format, size, owner, tags)
     VALUES (:id, :uri, :format, :size, :owner, :tags)
RETURNING created_at`

// SetBlobTags sets the tag list for a blob.
func (pg *PGClient) SetBlobTags(id string, tags []string) error {
	result, err := pg.DB.Exec(`UPDATE blobs SET tags = $1 WHERE ID = $2`,
		chassis.Tags(tags), id)
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ErrBlobNotFound
	}
	return nil
}

// ClearBlobOwner deletes the given blob from the image gallery of its
// owning user, returning a boolean flag indicating whether the blob
// is still in use in association with items.
func (pg *PGClient) ClearBlobOwner(id string) (bool, error) {
	tx, err := pg.DB.Beginx()
	if err != nil {
		return false, err
	}
	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			tx.Rollback()
		}
	}()

	result, err := tx.Exec(`UPDATE blobs SET owner = NULL WHERE ID = $1`,
		id)
	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	if rows != 1 {
		return false, ErrBlobNotFound
	}
	return blobInUse(tx, id)
}

// DeleteBlob performs the actual deletion of a blob and any item
// associations (which are cleaned up by a foreign key deletion
// cascade constraint).
func (pg *PGClient) DeleteBlob(id string) error {
	result, err := pg.DB.Exec(`DELETE FROM blobs WHERE ID = $1`,
		id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ErrBlobNotFound
	}
	return nil
}

// TagsForUser returns all the tags used in blobs owned by a user.
func (pg *PGClient) TagsForUser(userID string) ([]string, error) {
	tags := chassis.Tags{}
	err := pg.DB.Get(&tags,
		`SELECT jsonb_object_agg(tags) FROM blobs WHERE owner = $1`, userID)
	if err != nil {
		return nil, err
	}
	return tags, nil
}

// AddBlobsToItem adds associations between a list of blobs and an
// item.
func (pg *PGClient) AddBlobsToItem(itemID string, blobIDs []string) error {
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

	stmt, err := tx.Preparex(addBlobsToItem)
	for _, blobID := range blobIDs {
		_, err := stmt.Exec(blobID, itemID)
		if err != nil {
			return err
		}
	}
	return err
}

const addBlobsToItem = `
INSERT INTO blob_items (blob_id, item_id) VALUES ($1, $2)
  ON CONFLICT DO NOTHING`

// RemoveBlobsFromItem removes associations between a blob and a set
// of items (if the set passed in is empty, that means to delete all
// associations for the given item), deleting any resulting unowned
// blobs and returning information about the blobs that were deleted.
func (pg *PGClient) RemoveBlobsFromItem(itemID string, blobIDs []string) ([]DeletedBlob, error) {
	tx, err := pg.DB.Beginx()
	if err != nil {
		return nil, err
	}
	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			tx.Rollback()
		}
	}()

	deletedIDs := []string{}
	if len(blobIDs) > 0 {
		// Remove blob/item associations for blobs in the provided ID
		// list.
		query, args, err := sqlx.In(removeBlobsFromItem, itemID, blobIDs)
		if err != nil {
			return nil, err
		}
		query = tx.Rebind(query)
		err = tx.Select(&deletedIDs, query, args...)
	} else {
		// Remove all blob/item assocations for an item.
		err = tx.Select(&deletedIDs, removeAllBlobsFromItem, itemID)
	}
	if err != nil {
		return nil, err
	}

	// Determine which blobs are no longer in use.
	toDelete := []string{}
	for _, id := range deletedIDs {
		inuse, err := blobInUse(tx, id)
		if err != nil {
			return nil, err
		}
		if !inuse {
			toDelete = append(toDelete, id)
		}
	}
	if len(toDelete) == 0 {
		return []DeletedBlob{}, nil
	}

	// Delete unused blobs and collect the information needed to remove
	// their data from storage.
	results := []DeletedBlob{}
	query, args, err := sqlx.In(deleteUnusedBlobs, toDelete)
	if err != nil {
		return nil, err
	}
	query = tx.Rebind(query)
	err = tx.Select(&results, query, args...)
	if err != nil {
		return nil, err
	}

	return results, nil
}

const removeBlobsFromItem = `
DELETE FROM blob_items WHERE item_id = ? AND blob_id IN (?)
  RETURNING blob_id`

const removeAllBlobsFromItem = `
DELETE FROM blob_items WHERE item_id = $1
  RETURNING blob_id`

const deleteUnusedBlobs = `
DELETE FROM blobs WHERE id in (?)
RETURNING id, format`

// SaveEvent saves an event to the database.
func (pg *PGClient) SaveEvent(topic string, eventData interface{}, inTx func() error) error {
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
	err = chassis.SaveEvent(tx, topic, eventData, inTx)
	return err
}

// Determine whether a blob is still in use, either because it is
// owned by a user, or because it is associated with items.
func blobInUse(tx *sqlx.Tx, id string) (bool, error) {
	hasOwner := false
	err := tx.Get(&hasOwner,
		`SELECT owner IS NOT NULL FROM blobs WHERE id = $1`, id)
	if err != nil {
		return false, err
	}
	hasAssocs := false
	err = tx.Get(&hasAssocs,
		`SELECT COUNT(*) > 0 FROM blob_items WHERE blob_id = $1`, id)
	if err != nil {
		return false, err
	}
	return hasOwner || hasAssocs, nil
}
