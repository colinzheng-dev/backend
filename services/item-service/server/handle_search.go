package server

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/go-chi/chi"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/item-service/db"
	"github.com/veganbase/backend/services/item-service/model"
	"github.com/veganbase/backend/services/item-service/model/types"
)

// Handle normal item search, which allows filtering by item type,
// approval status (administrators only) and owner.
func (s *Server) itemSearch(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	return s.doSearch(w, r, "", "")
}

// Handle tag search, which allows the same filtering as normal
// search.
func (s *Server) tagSearch(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	tag := chi.URLParam(r, "tag")
	if tag == "" {
		return chassis.BadRequest(w, "missing search tag")
	}
	return s.doSearch(w, r, tag, "")
}

func (s *Server) itemsForOrg(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Get ID from URL parameters.
	idOrSlug := chi.URLParam(r, "id_or_slug")
	//getting org id by its id (validate if exists) or slug
	orgId, err := s.getOrgId(idOrSlug)
	if err != nil {
		return nil, err
	}

	return s.doSearch(w, r, "", *orgId)
}

func (s *Server) itemsForUser(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Get ID from URL parameters.
	paramUserID := chi.URLParam(r, "id")

	// Get authentication information from context.
	authInfo := chassis.AuthInfoFromContext(r.Context())

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

	return s.doSearch(w, r, "", actionUserID)
}

// Perform basic searches.
func (s *Server) doSearch(w http.ResponseWriter, r *http.Request, tag, owner string) (interface{}, error) {

	var userSession string
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod != chassis.NoAuth {
		userSession = authInfo.UserID
	}

	allowUser := true
	if owner != "" {
		allowUser = false
	}
	params, err := checkParams(r, allowUser)
	if err != nil {
		if err == ErrSearchNotFound {
			return chassis.NotFound(w)
		}
		return chassis.BadRequest(w, err.Error())
	}

	if owner != "" {
		params.Owner = &owner
	}

	var owners []string

	if owner == "" {
		owners, err = s.possibleOwners(owner)
		if err != nil {
			return nil, errors.New("possibleOwners: " + err.Error())
		}
	} else if len(owner) >= 4 {
		switch owner[0:4] {
		case "usr_":
			owners, err = s.possibleOwners(owner)
			if err != nil {
				return nil, errors.New("possibleOwners: " + err.Error())
			}
		case "org_":
			owners = append(owners, owner)
		default:
			return nil, errors.New("incorrect format in the owner field")
		}
	} else {
		return nil, errors.New("incorrect format in the owner field")
	}

	dbParams := db.SearchParams{
		ItemTypes: &params.ItemTypes,
		Approval:  params.Approval,
		Owner:     owners,
		Ids:       params.Ids,
		SortBy:    params.SortBy,
	}

	if tag != "" {
		dbParams.Tag = &tag
	}

	var idFilter []string
	s.getFilterIDs(params, &idFilter)
	if idFilter != nil && len(idFilter) == 0 {
		return []model.Item{}, nil
	}
	fmt.Println("===> idFilter =", idFilter)

	var collectionIDs []string

	hasCollectionFilter := len(*params.Collections) > 0
	s.getCollectionIDs(params.Collections, &collectionIDs)
	fmt.Println("===> collectionIDs =", collectionIDs)
	if hasCollectionFilter && len(collectionIDs) == 0 {
		return chassis.NotFoundWithMessage(w, "no results found matching the params passed")
	}

	// Summary views are simple.
	if params.Format == SummaryResults {
		items, totalItems, err := s.db.SummaryItems(&dbParams, idFilter, collectionIDs, params.Pagination)
		if err != nil {
			return nil, errors.New("SummaryItems: " + err.Error())
		}

		if len(items) == 0 {
			return chassis.NotFoundWithMessage(w, "no results found matching the params passed")
		}

		chassis.BuildPaginationResponse(w, r, params.Pagination.Page, params.Pagination.PerPage, *totalItems)
		return items, nil
	}

	// For full item views, we also need to process user information and inter-item links.
	// Get raw item data and expand into views with user information.
	rawItems, totalItems, err := s.db.FullItems(&dbParams, idFilter, collectionIDs, params.Pagination)
	if err != nil {
		return nil, errors.New("FullItems: " + err.Error())
	}

	if len(rawItems) == 0 {
		return chassis.NotFoundWithMessage(w, "no results found matching the params passed")
	}

	items, err := s.ExpandItemViews(rawItems, userSession)
	if err != nil {
		return nil, errors.New("ExpandItemViews: " + err.Error())
	}
	chassis.BuildPaginationResponse(w, r, params.Pagination.Page, params.Pagination.PerPage, *totalItems)

	if params.Links == nil || len(params.Links) == 0 {
		return items, nil
	}

	// Process inter-item links.
	for _, item := range items {
		links, err := s.ProcessLinks(item, params.Links, userSession)
		if err != nil {
			return nil, errors.New("ProcessLinks: " + err.Error())
		}
		item.Links = links
	}
	return items, nil
}


func (s *Server) getFilterIDs(params *SearchParams, ids *[]string) {
	hasGeo := params.Geo != nil
	hasFullText := params.Q != nil

	if !hasGeo && !hasFullText {
		return
	}

	var wg sync.WaitGroup
	geoIDs := []string{}
	fullTextIDs := []string{}
	if hasGeo {
		wg.Add(1)
		go s.getGeoIDs(&wg, &geoIDs, params.Geo[0], params.Geo[1], *params.Dist)
	}
	if hasFullText {
		wg.Add(1)
		go s.getFullTextIDs(&wg, &fullTextIDs, *params.Q)
	}
	wg.Wait()

	if hasGeo && !hasFullText {
		*ids = geoIDs
		return
	}
	if !hasGeo && hasFullText {
		*ids = fullTextIDs
		return
	}

	// Intersect ID sets.
	geoIDMap := map[string]bool{}
	for _, id := range geoIDs {
		geoIDMap[id] = true
	}
	result := []string{}
	for _, id := range fullTextIDs {
		if _, ok := geoIDMap[id]; ok {
			result = append(result, id)
		}
	}
	*ids = result
}

func (s *Server) getGeoIDs(wg *sync.WaitGroup, ids *[]string, latitude, longitude, dist float64) {
	defer wg.Done()

	var err error
	*ids, err = s.searchSvc.Geo(latitude, longitude, dist)
	if err != nil {
		log.Error().Err(err).
			Msg("error return from search service for geo search")
	}
}

func (s *Server) getFullTextIDs(wg *sync.WaitGroup, ids *[]string, q string) {
	defer wg.Done()

	var err error
	*ids, err = s.searchSvc.FullText(q)
	if err != nil {
		log.Error().Err(err).
			Msg("error return from search service for full-text search")
	}
}

// Check search parameters for validity and implementation status.
func checkParams(r *http.Request, allowUser bool) (*SearchParams, error) {
	// Get authentication information from context to determine whether
	// user is an administrator.
	authInfo := chassis.AuthInfoFromContext(r.Context())

	params, err := Params(r, allowUser)
	if err != nil {
		return nil, err
	}

	// Linked items parameters.
	if params.Links != nil && params.Format != FullResults {
		return nil, errors.New("linked items can only be displayed with 'full' format")
	}

	// Approval filtering.
	if params.Approval != nil {
		// Non-admin users can only view their own non-approved items.
		requireAuth := false
		for _, app := range *params.Approval {
			if app != types.Approved {
				requireAuth = true
				break
			}
		}

		if requireAuth {
			if authInfo.AuthMethod == chassis.NoAuth {
				return nil, ErrSearchNotFound
			}

			if !authInfo.UserIsAdmin {
				if params.Owner != nil && *params.Owner != authInfo.UserID {
					return nil, errors.New("non-admin users are only allowed to view their own non-approved items")
				}
				if params.Owner == nil {
					params.Owner = &authInfo.UserID
				}
			}
		}
		return params, nil
	}

	return params, nil
}

func (s *Server) getCollectionIDs(collIDs *[]string, ids *[]string) {
	if collIDs == nil {
		return
	}
	itemIds := make(map[string]bool)
	for _, collection := range *collIDs {
		collView, err := s.db.CollectionViewByName(collection)
		if err != nil {
			continue
		}
		for _, it := range collView.IDs {
			itemIds[it] = true
		}
	}

	//getting unique items ids
	for k := range itemIds {
		*ids = append(*ids, k)
	}

}