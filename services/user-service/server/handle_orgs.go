package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/user-service/db"
	"github.com/veganbase/backend/services/user-service/events"
	"github.com/veganbase/backend/services/user-service/messages"
	"github.com/veganbase/backend/services/user-service/model"
)

const emailValidator = "^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$"

func (s *Server) createOrg(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Only administrators may create organisations.
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod != chassis.SessionAuth  {
		return chassis.NotFound(w)
	}

	// Read new organisation request body.
	body, err := chassis.ReadBody(r, 0)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	org := model.Organisation{}
	// TODO: VALIDATION AGAINST JSON SCHEMA
	if 	err = json.Unmarshal(body, &org); err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	var zeroTime time.Time
	if org.ID != "" {
		return chassis.BadRequest(w, "can't set read-only field 'id'")
	}
	if org.Slug != "" {
		return chassis.BadRequest(w, "can't set read-only field 'slug'")
	}
	if org.CreatedAt != zeroTime {
		return chassis.BadRequest(w, "can't set read-only field 'created_at'")
	}

	if err = s.db.CreateOrg(&org); err != nil {
		return nil, err
	}

	if err = s.db.OrgAddUser(org.ID, authInfo.UserID, true); err != nil {
		//if err == db.ErrOrgNotFound {
		//TODO: EVALUATE ERRORS ON PRODUCTION AND REMOVE THIS COMMENTED LINES ONCE TESTS ARE DONE
		fmt.Println("error calling OrgAddUser to org " + org.ID + ": ", err.Error())
		//}
	}

	return org, nil
}

func (s *Server) orgs(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	qs, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return chassis.BadRequest(w, "invalid query parameters")
	}
	search := qs.Get("q")
	var page, perPage uint
	if tmp, err := strconv.Atoi(qs.Get("page")); err == nil {
		page = uint(tmp)
	}
	if tmp, err := strconv.Atoi(qs.Get("per_page")); err == nil {
		perPage = uint(tmp)
	}
	orgs, err := s.db.Orgs(search, page, perPage)
	if err != nil {
		return nil, err
	}

	return orgs, nil
}

func (s *Server) orgDetail(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Get ID from URL parameters.
	idOrSlug := chi.URLParam(r, "id_or_slug")

	org, err := s.db.OrgByIDorSlug(idOrSlug)
	if err != nil {
		if err == db.ErrOrgNotFound {
			return chassis.NotFound(w)
		}
		return nil, err
	}

	return org, nil
}

func (s *Server) updateOrg(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Check whether the modification is allowed: user is administrator
	// or is organisation administrator for the organisation.
	org, allowed, err := s.orgModAllowed(w, r, false)
	if !allowed {
		return nil, err
	}

	// Read patch request body.
	body, err := chassis.ReadBody(r, 0)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	err = org.Patch(body)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	// Do the update.
	if err = s.db.UpdateOrg(org); err != nil {
		return nil, err
	}
	if err == db.ErrOrgNotFound {
		return chassis.NotFound(w)
	}
	chassis.Emit(s, events.OrgUpdated, org)
	s.Invalidate(org.ID)

	return org, nil
}

func (s *Server) deleteOrg(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Get authentication information from context and only allow
	// administrators to proceed.
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod != chassis.SessionAuth || !authInfo.UserIsAdmin {
		return chassis.NotFound(w)
	}

	// Get the organisation ID to update.
	idOrSlug := chi.URLParam(r, "id_or_slug")
	if idOrSlug == "" {
		return chassis.BadRequest(w, "missing organisation ID")
	}

	// Look up organisation value.
	org, err := s.db.OrgByIDorSlug(idOrSlug)
	if err != nil {
		if err == db.ErrOrgNotFound {
			return chassis.NotFound(w)
		}
		return nil, err
	}

	// Do the delete.
	err = s.db.DeleteOrg(org.ID)
	if err == db.ErrOrgNotFound {
		return chassis.NotFound(w)
	}
	if err != nil {
		return nil, err
	}
	chassis.Emit(s, events.OrgDeleted, org)
	s.Invalidate(org.ID)

	return chassis.NoContent(w)
}

// Blob list for a user. Has "owner or admin" semantics for
// /user/{userid}/orgs uses (also accessed via /me/orgs).
func (s *Server) userOrgs(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		return chassis.NotFound(w)
	}

	// Get ID from URL parameters.
	paramUserID := chi.URLParam(r, "id")

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

	orgs, err := s.db.UserOrgs(actionUserID)
	if err != nil {
		return nil, err
	}

	return orgs, nil
}

func (s *Server) orgUsers(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Get ID from URL parameters.
	idOrSlug := chi.URLParam(r, "id_or_slug")
	org, err := s.db.OrgByIDorSlug(idOrSlug)
	if err != nil {
		if err == db.ErrOrgNotFound {
			return chassis.NotFound(w)
		}
		return nil, err
	}

	ousers, err := s.db.OrgUsers(org.ID)
	if err == db.ErrOrgNotFound {
		return chassis.NotFound(w)
	}
	if err != nil {
		return nil, err
	}

	userIDs := map[string]bool{}
	for _, ou := range ousers {
		userIDs[ou.UserID] = true
	}
	uniqueUserIDs := []string{}
	for id := range userIDs {
		uniqueUserIDs = append(uniqueUserIDs, id)
	}
	userMap, err := s.db.UsersByIDs(uniqueUserIDs)
	if err != nil {
		return nil, err
	}

	result := []*model.UserWithOrgInfo{}
	for _, ou := range ousers {
		view := model.UserWithOrgAdminFlag(userMap[ou.UserID], ou.IsOrgAdmin)
		result = append(result, view)
	}

	return result, nil
}

func (s *Server) orgAddUser(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Check whether the modification is allowed: user is administrator
	// or is organisation administrator for the organisation.
	org, allowed, err := s.orgModAllowed(w, r, false)
	if !allowed {
		return nil, err
	}

	// Read user addition request body.
	body, err := chassis.ReadBody(r, 0)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}
	req := messages.OrgAddUser{}

	if err = json.Unmarshal(body, &req); err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	// create a new user if email field is not empty
	if req.Email != "" {
		if ok, err := regexp.MatchString(emailValidator, req.Email); err == nil && ok {
			user, new, err := s.db.LoginUser(req.Email, s.avatarGen)
			if err != nil {
				return nil, err
			}
			if new {
				//TODO: maybe we'll need to check site on header
				msg := messages.LoginRequest{
					Email:    user.Email,
					Site:     "veganlogin",
					Language: "en",
				}

				chassis.Emit(s, events.UserCreated, msg)
			}
			req.UserID = user.ID

		}
	}

	if err = s.db.OrgAddUser(org.ID, req.UserID, req.IsOrgAdmin); err != nil {
		if err == db.ErrOrgNotFound {
			return chassis.NotFound(w)
		}
		if err == db.ErrUserNotFound || err == db.ErrUserAlreadyInOrg {
			return chassis.BadRequest(w, err.Error())
		}
		return nil, err
	}

	s.Invalidate(req.UserID)

	return chassis.NoContent(w)
}

func (s *Server) orgPatchUser(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	org, allowed, err := s.orgModAllowed(w, r, false)
	if !allowed {
		return nil, err
	}

	userID := chi.URLParam(r, "user_id")
	if userID == "" {
		return chassis.BadRequest(w, "user ID missing")
	}

	// Read user patch request body.
	body, err := chassis.ReadBody(r, 0)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}
	req := messages.OrgPatchUser{}
	err = json.Unmarshal(body, &req)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	// Patch user.
	err = s.db.OrgPatchUser(org.ID, userID, req.IsOrgAdmin)
	if err == db.ErrOrgNotFound {
		return chassis.NotFound(w)
	}
	if err == db.ErrUserNotFound {
		return chassis.BadRequest(w, "user not a member of organisation")
	}
	if err != nil {
		return nil, err
	}
	s.Invalidate(userID)

	return chassis.NoContent(w)
}

func (s *Server) orgDeleteUser(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	org, allowed, err := s.orgModAllowed(w, r, true)
	if !allowed {
		return nil, err
	}

	userID := chi.URLParam(r, "user_id")
	if userID == "" {
		return chassis.BadRequest(w, "user ID missing")
	}

	// Remove user.
	err = s.db.OrgDeleteUser(org.ID, userID)
	if err == db.ErrUserNotFound {
		return chassis.BadRequest(w, "user not a member of organisation")
	}
	if err != nil {
		return nil, err
	}
	s.Invalidate(userID)

	return chassis.NoContent(w)
}

func (s *Server) orgModAllowed(w http.ResponseWriter,
	r *http.Request, userSelfAllowed bool) (*model.Organisation, bool, error) {
	// Get authentication information from context and only allow
	// authenticated users to proceed.
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		chassis.NotFound(w)
		return nil, false, nil
	}

	// Get the organisation ID to update.
	idOrSlug := chi.URLParam(r, "id_or_slug")
	if idOrSlug == "" {
		chassis.BadRequest(w, "missing organisation ID")
		return nil, false, nil
	}

	// Look up organisation value.
	org, err := s.db.OrgByIDorSlug(idOrSlug)
	if err != nil {
		if err == db.ErrOrgNotFound {
			chassis.NotFoundWithMessage(w, err.Error())
			return nil, false, nil
		}
		return nil, false, err
	}

	// If the user is not an administrator, they can only update
	// organisations for which they are administrators.
	allowed := authInfo.UserIsAdmin
	if !allowed {
		// Look up the organisation users.
		orgUsers, err := s.db.OrgUsers(org.ID)
		if err != nil {
			return nil, false, err
		}

		for _, user := range orgUsers {
			if user.UserID == authInfo.UserID && user.IsOrgAdmin {
				allowed = true
				break
			}
		}
	}

	// An exception is when users are allowed to modify their own
	// context within an organisation.
	if !allowed && userSelfAllowed {
		userID := chi.URLParam(r, "user_id")
		if userID == "" {
			chassis.BadRequest(w, "missing user ID")
			return nil, false, nil
		}
		allowed = userID == authInfo.UserID
	}

	if !allowed {
		chassis.Forbidden(w)
	}
	return org, allowed, nil
}
