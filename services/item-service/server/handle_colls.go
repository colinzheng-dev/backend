package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-chi/chi"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/item-service/db"
	"github.com/veganbase/backend/services/item-service/events"
	"github.com/veganbase/backend/services/item-service/model"
	"github.com/veganbase/backend/services/item-service/model/types"
)

func (s *Server) createColl(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Get authentication information from context and only allow
	// authenticated users to proceed.
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		return chassis.NotFound(w)
	}

	// Read request body.
	body, err := chassis.ReadBody(r, 0)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	// Unmarshal request data -- validates JSON request body.
	req := &model.ItemCollectionInfo{}
	err = json.Unmarshal(body, &req)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	// Set up collection ownership. Collections may be created belonging
	// to organisations or the creating user.
	if req.Owner == "" {
		req.Owner = authInfo.UserID
	} else {
		switch req.Owner[0:4] {
		case "usr_":
			if req.Owner != authInfo.UserID {
				return chassis.BadRequest(w, "cannot create collections owned by another user")
			}
		case "org_":
			check, err := s.userSvc.IsUserOrgMember(authInfo.UserID, req.Owner)
			if err != nil {
				return nil, err
			}
			if !check {
				return chassis.BadRequest(w, "not a member of owning organisation")
			}
		default:
			return chassis.BadRequest(w, "invalid 'owner' field")
		}
	}

	err = s.db.CreateCollection(req)
	if err == db.ErrItemCollectionNotFound ||
		err == db.ErrCollectionNameAlreadyExists {
		return chassis.BadRequest(w, err.Error())
	}
	if err != nil {
		return nil, err
	}
	return chassis.NoContent(w)
}

func (s *Server) listColls(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	var owners []string

	m, _ := url.ParseQuery(r.URL.RawQuery)
	ownerID := m.Get("owner")

	if ownerID != "" {
		owners = []string{ownerID}
	}

	return s.collViews(w, r, owners)
}

func (s *Server) collsForUser(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Get authentication information from context and only allow
	// authenticated users to proceed.
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		return chassis.NotFound(w)
	}

	// Get ID from URL parameters.
	paramUserID := chi.URLParam(r, "user_id")

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

	owners, err := s.possibleOwners(actionUserID)
	if err != nil {
		return nil, err
	}

	return s.collViews(w, r, owners)
}

func (s *Server) collsForOrg(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Get authentication information from context and only allow
	// authenticated users to proceed.
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		return chassis.NotFound(w)
	}

	orgID := chi.URLParam(r, "org_id")
	if orgID == "" {
		return chassis.BadRequest(w, "no organisation ID provided")
	}

	member, err := s.userSvc.IsUserOrgMember(authInfo.UserID, orgID)
	if err != nil {
		return nil, err
	}
	if !authInfo.UserIsAdmin && !member {
		return chassis.Forbidden(w)
	}

	return s.collViews(w, r, []string{orgID})
}

func (s *Server) collViews(w http.ResponseWriter, r *http.Request,
	owners []string) (interface{}, error) {
	qs, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return chassis.BadRequest(w, "invalid query parameters")
	}
	var page, perPage uint
	err = chassis.PaginationParams(qs, &page, &perPage)

	colls, total, err := s.db.CollectionViews(owners, page, perPage)
	if err != nil {
		return nil, err
	}
	chassis.BuildPaginationResponse(w, r, page, perPage, *total)
	return colls, nil
}

func (s *Server) collDetail(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Doesn't need authentication.

	collID := chi.URLParam(r, "coll_id")
	if collID == "" {
		return chassis.BadRequest(w, "no collection ID provided")
	}

	coll, err := s.db.CollectionViewByName(collID)
	if err == db.ErrItemCollectionNotFound {
		return chassis.BadRequest(w, err.Error())
	}
	if err != nil {
		return nil, err
	}

	return coll, nil
}

func (s *Server) deleteColl(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Get authentication information from context and only allow
	// authenticated users to proceed.
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		return chassis.Forbidden(w)
	}

	// Get the item collection ID to delete.
	collID := chi.URLParam(r, "coll_id")
	if collID == "" {
		return chassis.BadRequest(w, "missing item collection ID")
	}

	// Determine permission information for update.
	allowedOwners, err := s.allowedOwners(authInfo)
	if err != nil {
		return nil, err
	}

	// Delete the item collection.
	err = s.db.DeleteItemCollection(collID, allowedOwners)
	if err == db.ErrItemCollectionNotOwned {
		return chassis.Forbidden(w)
	}
	if err == db.ErrItemCollectionNotFound {
		return chassis.NotFound(w)
	}
	if err != nil {
		return nil, err
	}

	return chassis.NoContent(w)
}

// AddCollectionItem is the optional request body used for adding
// items to collections.
type AddCollectionItem struct {
	Before *string `json:"before"`
	After  *string `json:"after"`
}

func (s *Server) collAddItem(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Get authentication information from context and only allow
	// authenticated users to proceed.
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		return chassis.Forbidden(w)
	}

	// Get the item collection and item IDs.
	collID := chi.URLParam(r, "coll_id")
	if collID == "" {
		return chassis.BadRequest(w, "missing item collection ID")
	}
	itemID := chi.URLParam(r, "item_id")
	if itemID == "" {
		return chassis.BadRequest(w, "missing item ID")
	}

	// Read optional request body.
	body, err := chassis.ReadBody(r, 1)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}
	addParams := AddCollectionItem{}
	if len(body) > 0 {
		err = json.Unmarshal(body, &addParams)
		if err != nil {
			return chassis.BadRequest(w, err.Error())
		}
	}

	// Determine permission information for update.
	allowedOwners, err := s.allowedOwners(authInfo)
	if err != nil {
		return nil, err
	}

	// Add the item to the item collection.
	err = s.db.AddItemToCollection(collID, itemID,
		addParams.Before, addParams.After, allowedOwners)
	if err == db.ErrItemCollectionNotOwned {
		return chassis.Forbidden(w)
	}
	if err == db.ErrItemCollectionNotFound {
		msgError := fmt.Sprintf("%s is not found", collID)
		return chassis.NotFoundWithMessage(w, msgError)
	}
	if err != nil {
		return nil, err
	}

	s.emitWithCollection(events.ItemAddedToCollection, itemID, collID)

	return chassis.NoContent(w)
}

func (s *Server) collRemoveItem(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Get authentication information from context and only allow
	// authenticated users to proceed.
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		return chassis.Forbidden(w)
	}

	// Get the item collection and item IDs.
	collID := chi.URLParam(r, "coll_id")
	if collID == "" {
		return chassis.BadRequest(w, "missing item collection ID")
	}
	itemID := chi.URLParam(r, "item_id")
	if itemID == "" {
		return chassis.BadRequest(w, "missing item ID")
	}

	// Determine permission information for update.
	allowedOwners, err := s.allowedOwners(authInfo)
	if err != nil {
		return nil, err
	}

	// Delete the item from the item collection.
	err = s.db.DeleteItemFromCollection(collID, itemID, allowedOwners)
	if err == db.ErrItemCollectionNotOwned {
		return chassis.Forbidden(w)
	}
	if err == db.ErrItemCollectionNotFound {
		return chassis.NotFound(w)
	}
	if err != nil {
		return nil, err
	}

	s.emitWithCollection(events.ItemRemovedFromCollection, itemID, collID)

	return chassis.NoContent(w)
}

func (s *Server) collPreview(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	return nil, errors.New("NOT YET IMPLEMENTED")
}

type collListAllResponse struct {
	CollName string                  `json:"name"`
	Items    []*collListItemResponse `json:"items"`
}

type collListItemResponse struct {
	ID              string                `json:"id"`
	ItemType        model.ItemType        `json:"item_type"`
	Slug            string                `json:"slug"`
	Lang            string                `json:"lang"`
	Name            string                `json:"name"`
	Description     string                `json:"description"`
	FeaturedPicture string                `json:"featured_picture"`
	Pictures        pq.StringArray        `json:"pictures"`
	Tags            pq.StringArray        `json:"tags"`
	URLs            types.URLMap          `json:"urls"`
	Attrs           types.AttrMap         `json:"attrs"`
	Approval        types.ApprovalState   `json:"approval"`
	Creator         string                `json:"creator"`
	Owner           string                `json:"owner"`
	Ownership       types.OwnershipStatus `json:"ownership"`
	CreatedAt       time.Time             `json:"created_at"`
}

func (s *Server) collListAll(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Get authentication information from context and only allow
	// authenticated users to proceed.
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		return chassis.NotFoundWithMessage(w, "users is not authenticated")
	}

	owner := r.URL.Query().Get("owner")

	// Get ID from URL parameters.
	paramUserID := chi.URLParam(r, "user_id")

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
		return chassis.NotFoundWithMessage(w, "user is not admin")
	}

	//get all orgs that the user is related
	owners, err := s.possibleOwners(actionUserID)
	if err != nil {
		return chassis.NotFoundWithMessage(w, err.Error())
	}

	// if owner was passed as query param, check if it belongs to  user's orgs
	// if not, returns an error
	if owner != "" {
		isOwner := false
		for _, id := range owners {
			if id == owner {
				isOwner = true
				owners = []string{id}
				break
			}
		}
		if !isOwner {
			return chassis.NotFoundWithMessage(w, "authenticated user is not related to this org")
		}
	}

	response := []*collListAllResponse{}

	collItens, err := s.db.CollectionViewsByOwners(owners)
	if err != nil {
		return chassis.NotFoundWithMessage(w, err.Error())
	}

	for _, c := range collItens {
		coll := &collListAllResponse{}
		coll.CollName = c.Name

		items, err := s.db.FullItemsByIDs(c.IDs)
		if err != nil {
			continue
		}

		for _, item := range items {
			collItem := new(collListItemResponse)
			collItem.ID = item.ID
			collItem.ItemType = item.ItemType
			collItem.Slug = item.Slug
			collItem.Lang = item.Lang
			collItem.Name = item.Name
			collItem.Description = item.Description
			collItem.FeaturedPicture = item.FeaturedPicture
			collItem.Pictures = item.Pictures
			collItem.Tags = item.Tags
			collItem.URLs = item.URLs
			collItem.Attrs = item.Attrs
			collItem.Approval = item.Approval
			collItem.Creator = item.Creator
			collItem.Owner = item.Owner
			collItem.Ownership = item.Ownership
			collItem.CreatedAt = item.CreatedAt

			coll.Items = append(coll.Items, collItem)
		}

		response = append(response, coll)
	}

	return response, nil
}
