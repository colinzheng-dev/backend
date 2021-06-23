package server

import (
	"github.com/go-chi/chi"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/user-service/db"
	"github.com/veganbase/backend/services/user-service/events"
	"github.com/veganbase/backend/services/user-service/model"
)

func (s *Server) list(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	_, _, isAdmin := accessControl(r)
	if !isAdmin {
		return chassis.NotFound(w)
	}

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
	users, err := s.db.Users(search, page, perPage)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (s *Server) detail(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	userID, fullProfile := detailAccessControl(r)
	if userID == nil {
		return chassis.NotFound(w)
	}

	user, err := s.db.UserByID(*userID)
	if err == db.ErrUserNotFound {
		return chassis.NotFound(w)
	}
	if err != nil {
		return nil, err
	}
	orgs, err := s.db.UserOrgs(*userID)
	if err != nil {
		return nil, err
	}

	if fullProfile {
		profile := model.UserFullProfile{User: *user}
		if len(orgs) > 0 {
			profile.Orgs = &orgs
		}
		return profile, nil
	}
	profile := model.UserPublicProfile{
		ID:          user.ID,
		DisplayName: user.DisplayName,
		Avatar:      user.Avatar,
		Country:     user.Country,
	}
	if len(orgs) > 0 {
		profile.Orgs = &orgs
	}
	return profile, nil
}

func (s *Server) info(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	qs, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return chassis.BadRequest(w, "invalid query parameters")
	}

	idparam := qs.Get("ids")
	if idparam == "" {
		return chassis.BadRequest(w, "invalid user ID list")
	}

	ids := strings.Split(idparam, ",")
	info, err := s.db.Info(ids)
	if err != nil {
		return nil, err
	}

	return info, nil
}

func (s *Server) delete(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	userID, _, _ := accessControl(r)
	if userID == nil {
		return chassis.NotFound(w)
	}

	err := s.db.DeleteUser(*userID)
	if err == db.ErrUserNotFound {
		return chassis.NotFound(w)
	}
	if err != nil {
		return nil, err
	}
	chassis.Emit(s, events.UserDeleted, map[string]string{"user_id": *userID})
	s.Invalidate(*userID)
	w.WriteHeader(http.StatusNoContent)
	return nil, nil
}

func (s *Server) update(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	userID, actingUserID, adminAction := accessControl(r)
	if userID == nil {
		return chassis.NotFound(w)
	}

	// Read patch request body.
	body, err := chassis.ReadBody(r, 0)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	// Look up user value and patch it.
	user, err := s.db.UserByID(*userID)
	if err != nil {
		return nil, err
	}

	if 	err = user.Patch(body, *actingUserID, adminAction); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return nil, nil
	}

	if err = s.db.UpdateUser(user); err != nil {
		return nil, err
	}
	chassis.Emit(s, events.UserUpdated, user)
	s.Invalidate(*userID)

	return user, nil
}

//getNotificationInfoInternal is a method to return sensitive information (such as email and name)
// without authentication. This method is to provide information internally.
func (s *Server) getNotificationInfoInternal(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	id := chi.URLParam(r, "id")
	if id == ""{
		return chassis.BadRequest(w, "user or org id is missing")
	}

	if id[:4] == "usr_" {
		return s.db.NotificationInfoByUserId(id)
	}

	return s.db.NotificationInfoByOrgId(id)

}

func (s *Server) getUserByApiKeyInternal(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	key := chi.URLParam(r, "key")
	if key == "" {
		return chassis.BadRequest(w, "key is missing")
	}

	qs, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return chassis.BadRequest(w, "invalid query parameters")
	}

	secret := qs.Get("secret")
	if secret == "" {
		return chassis.BadRequest(w, "secret is missing")
	}

	//getting user by key
	user, err := s.db.UserByAPIKey(key)
	if err != nil {
		return nil, err
	}

	//validating if secret matches
	if chassis.CompareHashedKeys(*user.APISecretKey, secret) {
		user.APISecretKey = nil
		return user, err
	}
	return chassis.NotFound(w)
}