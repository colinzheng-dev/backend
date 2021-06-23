package server

import (
	"github.com/go-chi/chi"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/purchase-service/db"
	"github.com/veganbase/backend/services/purchase-service/events"
	"net/http"
	"net/url"
)

// get all subscription items of a specific owner or a specific ID
func (s *Server) subscriptionItemSearch(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		return chassis.NotFound(w)
	}

	subID := chi.URLParam(r, "sub_id")

	qs, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return chassis.BadRequest(w, "error while parsing the query")
	}

	params := chassis.Pagination{}
	// Pagination parameters.
	if err := chassis.PaginationParams(qs, &params.Page, &params.PerPage); err != nil {
		return nil, err
	}

	if subID != "" {
		sub, err := s.db.SubscriptionItemById(subID)
		if err != nil {
			if err == db.ErrSubscriptionItemNotFound {
				return chassis.NotFoundWithMessage(w, "subscription item not found")
			}
			return nil, err
		}
		if authInfo.UserID != sub.Owner {
			return chassis.NotFound(w)
		}
	}


	subs, total, err := s.db.SubscriptionItemsByOwner(authInfo.UserID, &params)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}
	chassis.BuildPaginationResponse(w, r, params.Page, params.PerPage, *total)
	return subs, nil

}

// patchSubscriptionItem performs a patch operation in a subscription item
func (s *Server) patchSubscriptionItem(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		return chassis.NotFound(w)
	}

	subID := chi.URLParam(r, "sub_id")

	// Read patch request body.
	body, err := chassis.ReadBody(r, 0)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	// check if subscription item exists on the database
	sub, err := s.db.SubscriptionItemById(subID)
	if err != nil {
		if err == db.ErrOrderNotFound {
			return chassis.NotFoundWithMessage(w, "subscription item not found")
		}
		return nil, err
	}

	//check if authenticated user is owner of the subscription item
	if authInfo.UserID != sub.Owner {
		return chassis.BadRequest(w, "user authenticated is not the owner of subscription item")
	}


	if err = sub.Patch(body); err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	// Do the update.
	if err = s.db.UpdateSubscriptionItem(sub); err != nil {
		return nil, err
	}

	chassis.Emit(s, events.SubscriptionItemUpdated, sub)

	return sub, nil
}


// flipStateSubscriptionItem flips the state between 'paused' and 'active'
func (s *Server) flipStateSubscriptionItem(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		return chassis.NotFound(w)
	}

	subID := chi.URLParam(r, "sub_id")

	// check if subscription item exists on the database
	sub, err := s.db.SubscriptionItemById(subID)
	if err != nil {
		if err == db.ErrOrderNotFound {
			return chassis.NotFoundWithMessage(w, "subscription item not found")
		}
		return nil, err
	}

	//check if authenticated user is owner of the subscription item
	if authInfo.UserID != sub.Owner {
		return chassis.BadRequest(w, "user authenticated is not the owner of subscription item")
	}

	// Do the update.
	if err = s.db.FlipActivationState(sub); err != nil {
		return nil, err
	}

	chassis.Emit(s, events.SubscriptionItemUpdated, sub)

	return sub, nil
}

// deleteSubscriptionItem changes the state of the subscription item to deleted
func (s *Server) deleteSubscriptionItem(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		return chassis.NotFound(w)
	}

	subID := chi.URLParam(r, "sub_id")

	// check if subscription item exists on the database
	sub, err := s.db.SubscriptionItemById(subID)
	if err != nil {
		if err == db.ErrOrderNotFound {
			return chassis.NotFoundWithMessage(w, "subscription item not found")
		}
		return nil, err
	}

	//check if authenticated user is owner of the subscription item
	if authInfo.UserID != sub.Owner {
		return chassis.BadRequest(w, "user authenticated is not the owner of subscription item")
	}


	// Do the update.
	if err = s.db.DeleteSubscriptionItem(sub); err != nil {
		return nil, err
	}

	chassis.Emit(s, events.SubscriptionItemDeleted, sub)

	return sub, nil
}
