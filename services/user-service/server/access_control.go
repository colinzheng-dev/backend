package server

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/veganbase/backend/chassis"
)

// accessControl applies access control rules to user routes. These
// basically implement "session authentication only" + "owner or
// admin" access rules for all access to user account information.
// Returns the user ID being operated on, the user ID of the user
// accessing the route and the admin flag of the user accessing the
// route.

// TODO: MAKE THIS A MIDDLEWARE.
func accessControl(r *http.Request) (*string, *string, bool) {
	// Get ID from URL parameters.
	paramUserID := chi.URLParam(r, "id")

	// Get authentication information from context.
	authInfo := chassis.AuthInfoFromContext(r.Context())

	// Logic:

	// Only session authentication is allowed for these routes.
	if authInfo.AuthMethod != chassis.SessionAuth {
		return nil, nil, false
	}

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
		return nil, nil, false
	}

	// Operation is allowed: return the user ID we're operating on.
	return &actionUserID, &authInfo.UserID, authInfo.UserIsAdmin
}

// The user detail view has special access control rules, since we
// need to return a public profile for user accesses that might
// otherwise be forbidden.
func detailAccessControl(r *http.Request) (*string, bool) {
	// Get ID from URL parameters.
	paramUserID := chi.URLParam(r, "id")

	// Get authentication information from context.
	authInfo := chassis.AuthInfoFromContext(r.Context())

	// Logic:

	// Any type of authentication is allowed for these routes.

	// The user ID we are trying to operate on is either from the
	// authenticated user, or from the URL parameter if it's there.
	actionUserID := authInfo.UserID
	if paramUserID != "" {
		actionUserID = paramUserID
	}

	// If the user ID that we're trying to operate on is different from
	// the requesting authenticated user ID, then we will only return
	// public information.
	fullProfile := true
	if actionUserID != authInfo.UserID && !authInfo.UserIsAdmin {
		fullProfile = false
	}

	// Operation is allowed: return the user ID we're operating on and
	// whether or not we should return a full profile.
	return &actionUserID, fullProfile
}
