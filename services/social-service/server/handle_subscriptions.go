package server

import (
	"github.com/go-chi/chi"
	"github.com/veganbase/backend/chassis"
	"net/http"
)

func (s *Server) listSubscriptions(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Get authentication information from context and only allow
	// authenticated users to proceed.
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		w.WriteHeader(http.StatusForbidden)
		return nil,nil
	}

	subscriptions, err := s.db.ListUserSubscriptions(authInfo.UserID)

	if err != nil {
		return nil, err
	}

	var data = map[string][]string{}

	data["subscriptions"] = subscriptions

	return data, nil
}

func (s *Server) listFollowers(w http.ResponseWriter, r *http.Request) (interface{}, error) {

	subscriptionID := chi.URLParam(r, "subscription_id")

	followers, err := s.db.ListFollowers(subscriptionID)

	if err != nil {
		return nil, err
	}

	var data = map[string][]string{}

	data["followers"] = followers

	return data, nil
}

func (s *Server) addSubscription(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Get authentication information from context and only allow
	// authenticated users to proceed.
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		w.WriteHeader(http.StatusForbidden)
		return nil,nil
	}

	subscriptionID := chi.URLParam(r, "subscription_id")

	err := s.db.CreateUserSubscription(authInfo.UserID, subscriptionID)

	if err != nil {
		return nil, err
	}

	w.WriteHeader(http.StatusCreated)

	return nil, nil
}

func (s *Server) deleteSubscription(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Get authentication information from context and only allow
	// authenticated users to proceed.
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		w.WriteHeader(http.StatusForbidden)
		return nil,nil
	}

	subscriptionID := chi.URLParam(r, "subscription_id")

	err := s.db.DeleteUserSubscription(authInfo.UserID, subscriptionID)

	if err != nil {
		return nil, err
	}

	w.WriteHeader(http.StatusNoContent)

	return nil, nil
}
