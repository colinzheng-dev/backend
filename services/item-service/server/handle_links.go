package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-chi/chi"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/item-service/db"
	"github.com/veganbase/backend/services/item-service/messages"
)

func (s *Server) createLink(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Get authentication information from context and only allow
	// authenticated users to proceed.
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		return chassis.NotFound(w)
	}

	// Get origin item ID.
	originID := chi.URLParam(r, "item_id")

	// Read request body.
	body, err := chassis.ReadBody(r, 0)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	// Unmarshal request data.
	req := messages.CreateLinkRequest{}
	err = json.Unmarshal(body, &req)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	// If the user is not an administrator, they can only create links
	// between items that they or organisations they're members of own.
	allowedOwners, err := s.allowedOwners(authInfo)
	if err != nil {
		return nil, err
	}

	// Create link.
	link, err := s.db.CreateLink(req.LinkType, originID, req.Target,
		authInfo.UserID, allowedOwners)
	if err == db.ErrItemNotFound {
		return chassis.NotFound(w)
	}
	if err != nil {
		return nil, err
	}

	return link, nil
}

func (s *Server) deleteLink(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Get authentication information from context and only allow
	// authenticated users to proceed.
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		return chassis.NotFound(w)
	}

	// Get link ID.
	linkID := chi.URLParam(r, "link_id")
	if linkID == "" {
		return chassis.BadRequest(w, "missing link ID")
	}

	// If the user is not an administrator, they can only delete links
	// between items that they or organisations they're members of own.
	allowedOwners, err := s.allowedOwners(authInfo)
	if err != nil {
		return nil, err
	}

	// Delete the link.
	err = s.db.DeleteLink(linkID, authInfo.UserID, allowedOwners)
	if err == db.ErrLinkOwnershipInvalid {
		return chassis.Forbidden(w)
	}
	if err == db.ErrLinkNotFound {
		return chassis.NotFound(w)
	}
	if err != nil {
		return nil, err
	}
	// s.emit(events.ItemLinkDeleted, linkID)

	return chassis.NoContent(w)
}

func (s *Server) getItemLinks(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Get origin item ID.
	itemID := chi.URLParam(r, "item_id")
	if itemID == "" {
		return chassis.BadRequest(w, "missing item ID")
	}

	// Pagination parameters.
	qs, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return chassis.BadRequest(w, "invalid query parameters")
	}
	var page, perPage uint
	if err := chassis.PaginationParams(qs, &page, &perPage); err != nil {
		return chassis.BadRequest(w, "invalid pagination parameters")
	}

	// Retrieve the links.
	links, total, err := s.db.LinksByOriginID(itemID, page, perPage)
	if err != nil {
		if err == db.ErrItemNotFound {
			errMsg := fmt.Sprintf("itemID: %s, page: %d, perPage: %d, err: %s.", itemID, page, perPage, err.Error())
			return chassis.NotFoundWithMessage(w, errMsg)
		}
		return nil, err
	}
	chassis.BuildPaginationResponse(w, r, page, perPage, *total)

	return links, nil
}

func (s *Server) getItemLink(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Get link ID.
	linkID := chi.URLParam(r, "link_id")
	if linkID == "" {
		return chassis.BadRequest(w, "missing link ID")
	}

	// Retrieve the link.
	link, err := s.db.LinkByID(linkID)
	if err != nil {
		if err == db.ErrLinkNotFound {
			return chassis.NotFound(w)
		}
		return nil, err
	}

	return link, nil
}
