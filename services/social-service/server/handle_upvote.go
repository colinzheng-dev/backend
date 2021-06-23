package server

import (
	"github.com/go-chi/chi"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/social-service/db"
	"github.com/veganbase/backend/services/social-service/events"
	"github.com/veganbase/backend/services/social-service/model"
	"net/url"
	"strings"

	"net/http"
)

//TODO: MIGRATE THESE PATH TO ITEM-SERVICE TO ALLOW FILTERING
func (s *Server) getUserUpvotesProtected(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		return chassis.Unauthorized(w, "this resource can only be accessed by authorized users")
	}
	return s.getUserUpvotes(w, r)
}

func (s *Server) getUserUpvotes(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())

	// Get user ID from URL parameters.
	userId := chi.URLParam(r, "user_id")
	if userId == "" {
		userId = authInfo.UserID
	}

	itemsByUser, err := s.db.ListUserUpvotes(userId)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}
	//todo: call item service to get all items information

	return s.itemSvc.GetItems(*itemsByUser, "")

}

func (s *Server) getItemUpvotesQuantities(w http.ResponseWriter, r *http.Request) (interface{}, error) {

	// Get user ID from URL parameters.
	qs, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return chassis.BadRequest(w, "invalid query parameters")
	}
	var ids []string
	rawIds := qs.Get("ids")

	if rawIds != "" {
		ids = strings.Split(rawIds, ",")
	}

	userSession := qs.Get("user")

	quantityInfo, err := s.db.UpvoteQuantityByItemId(ids)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	info := make(map[string]model.UpvoteQuantityInfo)

	for _, i := range *quantityInfo {
		info[i.ItemId] = i
	}

	if userSession != "" {
		userUpvotes, err := s.db.ListUserUpvotes(userSession)
		if err != nil {
			return chassis.BadRequest(w, err.Error())
		}

		for _, upvotedItem := range *userUpvotes {
			if _, ok := info[upvotedItem]; ok {
				updated := info[upvotedItem]
				updated.UserUpvoted = true
				info[upvotedItem] = updated
			}
		}
	}

	return info, err
}


func (s *Server) getItemsUpvotedIds(w http.ResponseWriter, r *http.Request) (interface{}, error) {

	userId := chi.URLParam(r, "user_id")
	if userId == "" {
		chassis.BadRequest(w,"missing user_id")
	}


	upvoted := make(map[string]bool)

	userUpvotes, err := s.db.ListUserUpvotes(userId)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	for _, upvotedItem := range *userUpvotes {
		upvoted[upvotedItem] = true
	}

	return upvoted, err
}

func (s *Server) createUpvote(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Get authentication information from context
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		return chassis.NotFound(w)
	}
	itemId := chi.URLParam(r, "item_id")

	if 	_, err := s.itemSvc.ItemInfo(itemId); err != nil {
		return chassis.NotFoundWithMessage(w, err.Error())
	}

	vote := model.Upvote{
		UserId: authInfo.UserID,
		ItemId: itemId,
	}

	if err := s.db.CreateUpvote(&vote); err != nil {
		return chassis.BadRequest(w, "error while creating an upvote: " +err.Error())
	}

	chassis.Emit(s, events.UpvoteCreated, vote)
	return vote, nil
}


func (s *Server) deleteUpvote(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Get authentication information from context
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		return chassis.NotFound(w)
	}
	itemId := chi.URLParam(r, "item_id")

	if itemId == "" {
		return chassis.BadRequest(w, "missing item ID")
	}

	//check if upvote exists
	upvote, err := s.db.UpvoteByUserAndItemId(authInfo.UserID, itemId)
	if err != nil {
		if err == db.ErrUpvoteNotFound {
			return chassis.NotFoundWithMessage(w, "upvote not found")
		}
		return chassis.BadRequest(w, err.Error())
	}

	//deleting upvote

	if err = s.db.DeleteUpvote(upvote.Id); err != nil {
		return nil, err
	}
	chassis.Emit(s, events.UpvoteDeleted, upvote)

	return chassis.NoContent(w)
}


func (s *Server) getAllItemsUpvotesQuantities(w http.ResponseWriter, r *http.Request) (interface{}, error) {

	quantityInfo, err := s.db.UpvoteQuantityByItemId(nil)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	info := make(map[string]model.UpvoteQuantityInfo)

	for _, i := range *quantityInfo {
		info[i.ItemId] = i
	}

	return info, err
}