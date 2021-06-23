package server

import (
	"github.com/pkg/errors"
	"strings"
)

// Add associations between an entity and its blobs when the entity is created.
// a blob can be associated with a post or a reply, so these paths handle them
// as entities instead of making one for each type of struct.
func (s *Server) addBlobs(entityId string, pictures []string) error {
	if len(pictures) == 0 {
		return nil
	}

	blobs := s.urlsToBlobIDs(pictures)
	if err := s.blobSvc.AddItemBlobs(entityId, blobs); err != nil {
		return errors.Wrap(err, "failed to add blobs for new entity")
	}
	return nil
}

func (s *Server) updateBlobs(id string, before, after map[string]bool) error {
	if len(before) == 0 && len(after) == 0 {
		return nil
	}
	// Work out what blobs have been added or remove to the entity.
	add := []string{}
	for pic := range after {
		if _, ok := before[pic]; !ok {
			add = append(add, pic)
		}
	}
	remove := []string{}
	for pic := range before {
		if _, ok := after[pic]; !ok {
			remove = append(remove, pic)
		}
	}
	addBlobs := s.urlsToBlobIDs(add)
	removeBlobs := s.urlsToBlobIDs(remove)

	// Update the blob/entity associations.
	if len(addBlobs) > 0 {
		err := s.blobSvc.AddItemBlobs(id, addBlobs)
		if err != nil {
			return errors.Wrap(err, "failed to add blob associations for " + id)
		}
	}
	if len(removeBlobs) > 0 {
		err := s.blobSvc.RemoveItemBlobs(id, removeBlobs)
		if err != nil {
			return errors.Wrap(err, "failed to remove blob associations for " + id)
		}
	}
	return nil
}

// Remove blob/entity associations when the this entity is deleted.
func (s *Server) removeBlobs(id string, pics []string) error {
	if len(pics) == 0 {
		return nil
	}

	blobs := s.urlsToBlobIDs(pics)
	if err := s.blobSvc.RemoveItemBlobs(id, blobs); err != nil {
		return errors.Wrap(err, "failed to remove blob associations for "+ id)
	}
	return nil
}


// Convert image URLs to blob IDs by removing base URL and file
// extension. Skip URLs that don't live on the the image server.
func (s *Server) urlsToBlobIDs(urls []string) []string {
	result := []string{}
	for _, url := range urls {
		if !strings.HasPrefix(url, s.imageBaseURL) {
			continue
		}
		id := strings.TrimLeft(strings.TrimPrefix(url, s.imageBaseURL), "/")
		dot := strings.Index(id, ".")
		if dot >= 0 {
			id = id[:dot]
		}
		result = append(result, id)
	}
	return result
}