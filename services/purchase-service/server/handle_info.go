package server

import (
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/purchase-service/db"
	"net/http"
	"net/url"
)

//userPurchaseItem verifies if the user already purchase a certain item.
func (s *Server) userPurchaseItem(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Get purchase ID from URL parameters

	qs, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return chassis.BadRequest(w, "error while parsing the query")
	}

	userId := qs.Get("user_id")
	itemId := qs.Get("item_id")

	if userId == "" || itemId == "" {
		return chassis.BadRequest(w, "user_id and item_id must be passed")
	}

	purchases, _, err := s.db.PurchasesByOwner(userId, nil)
	if err != nil {
		if err == db.ErrPurchaseNotFound {
			return purchases, nil
		}
		return nil, err
	}

	for _, purchase := range *purchases {
		for _, item := range purchase.Items {
			if item.ItemId == itemId {
				return true, nil
			}
		}
	}
	return false, nil
}