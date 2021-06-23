package server

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi"
	"github.com/pkg/errors"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/item-service/db"
	"github.com/veganbase/backend/services/item-service/events"
	"github.com/veganbase/backend/services/item-service/model"
	"github.com/veganbase/backend/services/item-service/model/types"
)

func (s *Server) createItem(w http.ResponseWriter, r *http.Request) (interface{}, error) {
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
	item := model.Item{}
	err = item.UnmarshalJSON(body)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	// Items created by administrators are automatically approved. All
	// others need to be approved by an administrator.
	if authInfo.UserIsAdmin {
		item.Approval = types.Approved
	} else {
		item.Approval = types.Pending
	}

	// Set up item ownership. Items may be created belonging to
	// organisations or the creating user.
	item.Creator = authInfo.UserID
	item.Ownership = types.Creator
	if item.Owner == "" {
		item.Owner = authInfo.UserID
	} else {
		switch item.Owner[0:4] {
		case "usr_":
			if item.Owner != authInfo.UserID {
				return chassis.BadRequest(w, "cannot create items owned by another user")
			}
		case "org_":
			check, err := s.userSvc.IsUserOrgMember(authInfo.UserID, item.Owner)
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

	// Create the item.
	err = s.db.CreateItem(&item)
	if err != nil {
		return nil, err
	}
	s.emit(events.ItemCreated, item.ID)

	// Add item/blob associations for new item.
	err = s.addItemBlobs(&item)
	if err != nil {
		return nil, err
	}

	userInfo, err := s.userSvc.Info([]string{item.Creator, item.Owner})
	view := model.FullView(&item, userInfo[item.Creator], userInfo[item.Owner], 0, false, 0, []string{})
	return view, nil
}

// Add blob/item associations when item is created.
func (s *Server) addItemBlobs(it *model.Item) error {
	blobs := s.urlsToBlobIDs(it.Pictures)
	err := s.blobSvc.AddItemBlobs(it.ID, blobs)
	if err != nil {
		return errors.Wrap(err, "failed to add blobs for new item")
	}
	return nil
}

func (s *Server) getByIDOrSlug(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Get ID or slug from URL parameters.
	idOrSlug := chi.URLParam(r, "id_or_slug")

	var userSession string
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod != chassis.NoAuth {
		userSession = authInfo.UserID
	}

	// Parse any embedded link parameters.
	qs, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return nil, errors.New("invalid query parameters")
	}
	linkInfo, err := ParseLinkParams(qs)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	// Retrieve item.
	rawItem, err := s.db.ItemByIDOrSlug(idOrSlug)
	if err != nil {
		if err == db.ErrItemNotFound {
			return chassis.NotFound(w)
		}
		return nil, err
	}

	// Get item creator and owner.
	userInfo, err := s.userSvc.Info([]string{rawItem.Creator, rawItem.Owner})
	if err != nil {
		return nil, err
	}

	var userUpvotes *map[string]bool
	if userSession != "" {
		userUpvotes, _ = s.socialSvc.GetUserUpvotes(userSession)
	}

	collNames, err := s.db.CollectionsNamesByItemId([]string{rawItem.ID})
	if err != nil {
		return nil, err
	}

	var upvoteInfo bool
	if userUpvotes != nil {
		upvoteInfo = (*userUpvotes)[rawItem.ID]
	}

	item := model.FullView(&rawItem.Item, userInfo[rawItem.Creator], userInfo[rawItem.Owner], rawItem.Upvotes, upvoteInfo, rawItem.Rank, (*collNames)[rawItem.ID])

	// Process any inter-item links.
	if len(linkInfo) > 0 {
		links, err := s.ProcessLinks(item, linkInfo, userSession)
		if err != nil {
			return nil, err
		}
		item.Links = links
	}

	return item, nil
}

func (s *Server) patchItem(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Get authentication information from context and only allow
	// authenticated users to proceed.
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		return chassis.Forbidden(w)
	}

	// Get the item ID to update.
	itemID := chi.URLParam(r, "id")
	if itemID == "" {
		return chassis.BadRequest(w, "missing item ID")
	}

	// Read patch request body.
	body, err := chassis.ReadBody(r, 0)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	// Look up item value and patch it.
	item, err := s.db.ItemByID(itemID)
	if err == db.ErrItemNotFound {
		return chassis.NotFound(w)
	}
	if err != nil {
		return nil, err
	}
	picsBefore := map[string]bool{}
	for _, pic := range item.Pictures {
		picsBefore[pic] = true
	}
	err = item.Patch(body)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}
	picsAfter := map[string]bool{}
	for _, pic := range item.Pictures {
		picsAfter[pic] = true
	}

	// Determine permission information for update.
	allowedOwners, err := s.allowedOwners(authInfo)
	if err != nil {
		return nil, err
	}

	// Do the update.
	if err = s.db.UpdateItem(item, allowedOwners); err != nil {
		return nil, err
	}
	if err == db.ErrItemNotFound {
		return chassis.NotFound(w)
	}
	s.emit(events.ItemUpdated, item.ID)

	// Update the item/blob associations.
	err = s.updateItemBlobs(item.ID, picsBefore, picsAfter)
	if err != nil {
		return nil, err
	}

	return item, nil
}

func (s *Server) updateItemBlobs(itemID string, before, after map[string]bool) error {
	// Work out what blobs have been added or remove to the item.
	add := []string{}
	for pic := range after {
		if _, ok := before[pic]; !ok {
			add = append(add, pic)
		}
	}
	remove := []string{}
	for pic := range before {
		if _, ok := after[pic]; !ok {
			remove = append(remove, pic)
		}
	}
	addBlobs := s.urlsToBlobIDs(add)
	removeBlobs := s.urlsToBlobIDs(remove)

	// Update the blob/item associations.
	if len(addBlobs) > 0 {
		err := s.blobSvc.AddItemBlobs(itemID, addBlobs)
		if err != nil {
			return errors.Wrap(err, "failed to add blob associations for item")
		}
	}
	if len(removeBlobs) > 0 {
		err := s.blobSvc.RemoveItemBlobs(itemID, removeBlobs)
		if err != nil {
			return errors.Wrap(err, "failed to remove blob associations for item")
		}
	}
	return nil
}

func (s *Server) deleteItem(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Get authentication information from context and only allow
	// authenticated users to proceed.
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		return chassis.Forbidden(w)
	}

	// Get the item ID to delete.
	itemID := chi.URLParam(r, "id")
	if itemID == "" {
		return chassis.BadRequest(w, "missing item ID")
	}

	// Determine permission information for update.
	allowedOwners, err := s.allowedOwners(authInfo)
	if err != nil {
		return nil, err
	}

	// Delete the item.
	pics, err := s.db.DeleteItem(itemID, allowedOwners)
	if err != nil {
		if err == db.ErrItemNotOwned {
			return chassis.Forbidden(w)
		} else if err == db.ErrItemNotFound {
			return chassis.NotFound(w)
		}
		return nil, err
	}
	s.emit(events.ItemDeleted, itemID)

	err = s.removeItemBlobs(itemID, pics)
	if err != nil {
		return nil, err
	}
	return chassis.NoContent(w)
}

// Remove blob/item associations when item is deleted.
func (s *Server) removeItemBlobs(itemID string, pics []string) error {
	blobs := s.urlsToBlobIDs(pics)
	err := s.blobSvc.RemoveItemBlobs(itemID, blobs)
	if err != nil {
		return errors.Wrap(err, "failed to remove blob associations for item")
	}
	return nil
}

// Convert image URLs to blob IDs by removing base URL and file
// extension. Skip URLs that don't live on the the image server.
func (s *Server) urlsToBlobIDs(urls []string) []string {
	result := []string{}
	for _, url := range urls {
		if !strings.HasPrefix(url, s.imageBaseURL) {
			continue
		}
		id := strings.TrimLeft(strings.TrimPrefix(url, s.imageBaseURL), "/")
		dot := strings.Index(id, ".")
		if dot >= 0 {
			id = id[:dot]
		}
		result = append(result, id)
	}
	return result
}

//updateAvailability is a method used to perform changes in item available quantity.
func (s *Server) updateAvailability(w http.ResponseWriter, r *http.Request) (interface{}, error) {

	// Get the item ID to update.
	itemID := chi.URLParam(r, "id")
	if itemID == "" {
		return chassis.BadRequest(w, "missing item ID")
	}

	// Read patch request body.
	body, err := chassis.ReadBody(r, 0)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	// Look up item value.
	item, err := s.db.ItemByID(itemID)
	if err != nil {
		if err == db.ErrItemNotFound {
			return chassis.NotFound(w)
		}
		return nil, err
	}

	//patching availability information
	if err = item.UpdateAvailability(body); err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	// Do the update.
	if err = s.db.UpdateAvailability(item); err != nil {
		if err == db.ErrItemNotFound {
			return chassis.NotFound(w)
		}
		return nil, err
	}

	s.emit(events.ItemUpdated, item.ID)

	return item, nil
}
