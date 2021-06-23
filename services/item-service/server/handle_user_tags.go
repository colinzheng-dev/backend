package server

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/veganbase/backend/chassis"
)

func (s *Server) tagsForUser(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Get ID from URL parameters.
	paramUserID := chi.URLParam(r, "id")

	// Get authentication information from context.
	authInfo := chassis.AuthInfoFromContext(r.Context())

	// The user ID we are trying to operate on is either from the
	// authenticated user, or from the URL parameter if it's there.
	actionUserID := authInfo.UserID
	if paramUserID != "" {
		actionUserID = paramUserID
	}

	// If the user ID that we're trying to operate on is different from
	// the requesting authenticated user ID, then the user must be an
	// administrator.
	if actionUserID != authInfo.UserID && !authInfo.UserIsAdmin {
		return chassis.NotFound(w)
	}

	tags, err := s.db.TagsForUser(actionUserID)
	if err != nil {
		return nil, err
	}
	return tags, nil
}
