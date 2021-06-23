package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
	"github.com/h2non/filetype"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/blob-service/db"
	"github.com/veganbase/backend/services/blob-service/model"
)

var (
	allowedFileTypes = []string{"jpg", "png", "webp"}
)

// TODO: THERE IS IN GENERAL A LACK OF TRANSACTIONALITY IN HERE. NEED
// TO THINK OF A WAY TO DEAL WITH THIS SO THAT THE DATABASE INTERFACE
// CAN BE USED NICELY BOTH IN SIMPLE CALLS AND IN SEQUENCES OF CALLS
// THAT NEED TO RUN INSIDE A TRANSACTION.

// Create a blob from a file upload. Upload size is limited and the
// permitted file types are restricted.
// TODO: MAKE THIS A MULTIPART UPLOAD, GIVING A FILE AND A JSON BODY
// TO ASSIGN TAGS IN A SINGLE STEP.
func (s *Server) create(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		return chassis.NotFound(w)
	}

	// Read body, limiting to maximum upload size.
	data, err := chassis.ReadBody(r, 0)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	// Decode and check file type.
	fileType, err := filetype.Match(data)
	if err != nil {
		return chassis.BadRequest(w, "Can't determine file type")
	}
	allowed := false
	for _, t := range allowedFileTypes {
		if fileType.Extension == t {
			allowed = true
			break
		}
	}
	if !allowed {
		return chassis.BadRequest(w, "Unsupported file type")
	}

	// Write blob data to storage.
	id := s.db.NewBlobID()
	url, size, err := s.blobstore.Write(id, fileType.Extension, data)
	if err != nil {
		return nil, err
	}

	// Create database record for blob.
	blob := model.Blob{
		ID:     id,
		URI:    url,
		Format: fileType.Extension,
		Size:   int(size),
		Owner:  &authInfo.UserID,
	}
	err = s.db.CreateBlob(&blob)
	if err != nil {
		return nil, err
	}

	// Fix up URL to give image server URL instead of storage URL.
	blob.URI = s.blobURL(&blob)
	return &blob, nil
}

// Get details for a single blob.
func (s *Server) detail(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		return chassis.NotFound(w)
	}
	id := chi.URLParam(r, "id")
	blob, err := s.db.BlobByID(id)
	if err == db.ErrBlobNotFound {
		return chassis.NotFound(w)
	}
	if err != nil {
		return nil, err
	}
	if !(authInfo.UserIsAdmin ||
		blob.Owner != nil && *blob.Owner == authInfo.UserID) {
		return chassis.NotFound(w)
	}

	// Fix up URL to give image server URL instead of storage URL.
	blob.URI = s.blobURL(blob)
	return blob, nil
}

// Update a blob: the only allowed change is to modify the tags.
func (s *Server) update(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		return chassis.NotFound(w)
	}
	id := chi.URLParam(r, "id")
	blob, err := s.db.BlobByID(id)
	if err == db.ErrBlobNotFound {
		return chassis.NotFound(w)
	}
	if err != nil {
		return nil, err
	}
	if !(authInfo.UserIsAdmin || blob.Owner != nil && *blob.Owner == authInfo.UserID) {
		return chassis.NotFound(w)
	}

	// Read patch request body.
	type validPatch struct {
		Tags chassis.Tags `json:"tags"`
	}
	patch := validPatch{}
	dec := json.NewDecoder(r.Body)
	defer r.Body.Close()
	dec.DisallowUnknownFields()
	err = dec.Decode(&patch)
	if err != nil {
		return chassis.BadRequest(w, "invalid patch for blob")
	}

	// Do patch.
	err = s.db.SetBlobTags(id, patch.Tags)
	if err != nil {
		return nil, err
	}
	blob.Tags = patch.Tags
	// Fix up URL to give image server URL instead of storage URL.
	blob.URI = s.blobURL(blob)
	return blob, nil
}

// Delete a blob: the blob is only removed if it has no remaining item
// associations.
func (s *Server) delete(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		return chassis.NotFound(w)
	}

	id := chi.URLParam(r, "id")
	blob, err := s.db.BlobByID(id)
	if err == db.ErrBlobNotFound {
		return chassis.NotFound(w)
	}
	if err != nil {
		return nil, err
	}
	if !(authInfo.UserIsAdmin || blob.Owner != nil && *blob.Owner == authInfo.UserID) {
		return chassis.NotFound(w)
	}

	inUse, err := s.db.ClearBlobOwner(id)
	if err != nil {
		return nil, err
	}
	if inUse {
		// The blob still has item associations, so don't delete it for
		// real.
		return chassis.NoContent(w)
	}

	return s.deleteUnusedBlob(w, id, blob.Format)
}

// Delete an unused blob.
func (s *Server) deleteUnusedBlob(w http.ResponseWriter,
	id string, format string) (interface{}, error) {
	// Delete the blob from the blob storage and from the database.
	err := s.blobstore.Delete(id, format)
	if err != nil {
		return nil, err
	}
	err = s.db.DeleteBlob(id)
	if err != nil {
		return nil, err
	}

	return chassis.NoContent(w)
}

// Blob list for a user, possibly filtered by tags. Has "owner or
// admin" semantics for /user/{userid}/blobs uses (also accessed via
// /me/blobs).
func (s *Server) list(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		return chassis.NotFound(w)
	}

	// Get ID from URL parameters.
	paramUserID := chi.URLParam(r, "userid")

	// The user ID we are trying to operate on is either from the
	// authenticated user, or from the URL parameter if it's there.
	actionUserID := authInfo.UserID
	if paramUserID != "" {
		actionUserID = paramUserID
	}

	// If the user ID that we're trying to operate on is different from
	// the requesting authenticated user ID, then the user must be an
	// administrator and we must be using session authentication.
	if actionUserID != authInfo.UserID &&
		(!authInfo.UserIsAdmin || authInfo.AuthMethod != chassis.SessionAuth) {
		return chassis.NotFound(w)
	}

	qs, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return chassis.BadRequest(w, "invalid query parameters")
	}
	tagsstr := qs.Get("tags")
	var tags []string
	if tagsstr != "" {
		tags = strings.Split(tagsstr, ",")
	}
	var page, perPage uint
	if tmp, err := strconv.Atoi(qs.Get("page")); err == nil {
		page = uint(tmp)
	}
	if tmp, err := strconv.Atoi(qs.Get("per_page")); err == nil {
		perPage = uint(tmp)
	}
	blobs, err := s.db.BlobsByUser(actionUserID, tags, page, perPage)
	if err != nil {
		return nil, err
	}

	// Fix up blob URLs.
	for i := range blobs {
		blobs[i].URI = s.blobURL(&blobs[i])
	}
	return blobs, nil
}

// Generate the public download URL for a blob (as opposed to the
// private Google Storage URL, which is what we store in the
// database).
func (s *Server) blobURL(blob *model.Blob) string {
	return fmt.Sprintf("%s/%s.%s", s.imageBaseURL, blob.ID, blob.Format)
}
