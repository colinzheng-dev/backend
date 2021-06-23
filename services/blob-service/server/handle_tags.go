package server

import (
	"net/http"

	"github.com/veganbase/backend/chassis"
)

// GET /blobs/tags
func (s *Server) tagList(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth || authInfo.UserID == "" {
		return chassis.NotFound(w)
	}

	tags, err := s.db.TagsForUser(authInfo.UserID)
	if err != nil {
		return nil, err
	}

	return tags, nil
}
