package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/pkg/errors"
	"github.com/veganbase/backend/chassis"
)

// addItemAssoc adds an association between an item and a blob. This
// is used when items are created or modified to make sure that a
// record is kept of the blobs that an item uses, so that they don't
// get deleted prematurely.
//
// Permissions here: the images associated with blobs are publicly
// accessible, so there's no problem with associating a blob with any
// item. Additionally, this route is not accessible via the API
// gateway, and authorisation for the item is handled by the item
// service.
func (s *Server) addItemBlobs(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Get item ID.
	itemID := chi.URLParam(r, "item")

	// Read body, limiting to maximum upload size.
	data, err := chassis.ReadBody(r, 0)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}
	blobIDs := []string{}
	err = json.Unmarshal(data, &blobIDs)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshalling blob ID list")
	}

	// Add blob/item associations.
	if err = s.db.AddBlobsToItem(itemID, blobIDs); err != nil {
		return nil, err
	}
	return chassis.NoContent(w)
}

// removeItemBlobs removes associations between an item and a blob.
// This is used when items are created or modified, and triggers
// deletion of blobs that are no longer in use (i.e. are not
// associated with any items and no longer have an owner because their
// owner deleted them). An empty request body triggers removal of all
// blob associations for an item.
//
// Permissions here: this route is not accessible via the API gateway,
// and permissions to modify the associated item are handled by the
// item service.
func (s *Server) removeItemBlobs(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Get item ID.
	itemID := chi.URLParam(r, "item")

	// Read body, allowing an empty body...
	data, err := chassis.ReadBody(r, 1)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}
	blobIDs := []string{}
	if len(data) > 0 {
		err = json.Unmarshal(data, &blobIDs)
		if err != nil {
			return nil, errors.Wrap(err, "unmarshalling blob ID list")
		}
	}

	// Remove blob/item associations.
	deleted, err := s.db.RemoveBlobsFromItem(itemID, blobIDs)
	if err != nil {
		return nil, err
	}

	// Remove blobs that have been deleted from the database from
	// storage.
	for _, d := range deleted {
		err := s.blobstore.Delete(d.ID, d.Format)
		if err != nil {
			return nil, err
		}
	}

	return chassis.NoContent(w)
}
