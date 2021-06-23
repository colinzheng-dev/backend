package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"sync"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog/log"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/item-service/db"
	"github.com/veganbase/backend/services/item-service/model"
	"github.com/veganbase/backend/services/item-service/model/types"
)

func (s *Server) claimItemOwnership(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Get authentication information from context and only allow
	// authenticated users to proceed.
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		return chassis.NotFound(w)
	}

	// Look up item.
	id := chi.URLParam(r, "id")
	_, err := s.db.ItemByID(id)
	if err == db.ErrItemNotFound {
		return chassis.NotFound(w)
	}
	if err != nil {
		return nil, err
	}

	// Check for a request body (used for organisation ownership
	// claims). (Allow empty request body for user claims.)
	ownerID := authInfo.UserID
	body, err := chassis.ReadBody(r, 1)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}
	if len(body) > 0 {
		orgBody := struct {
			OrgID string `json:"org"`
		}{}
		err = json.Unmarshal(body, &orgBody)
		if err != nil {
			return chassis.BadRequest(w, err.Error())
		}
		member, err := s.userSvc.IsUserOrgMember(authInfo.UserID, orgBody.OrgID)
		if err != nil {
			return nil, err
		}
		if !member {
			return chassis.Forbidden(w)
		}
		ownerID = orgBody.OrgID
	}

	// Create claim.
	claim := &model.Claim{
		OwnerID: ownerID,
		ItemID:  id,
	}
	err = s.db.CreateClaim(claim)
	if err != nil {
		return nil, err
	}
	return claim, nil
}

func (s *Server) listOwnershipClaims(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Get authentication information from context and only allow
	// authenticated users to proceed.
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		return chassis.NotFound(w)
	}

	qs, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return chassis.BadRequest(w, "invalid query parameters")
	}

	// Pagination parameters.
	var page, perPage uint
	if err := chassis.PaginationParams(qs, &page, &perPage); err != nil {
		return chassis.BadRequest(w, "invalid pagination parameters")
	}

	// Filtering parameters.
	status := (*types.ApprovalState)(nil)
	p := qs.Get("status")
	if p != "" {
		a := types.ApprovalState(types.Pending)
		if err := a.FromString(p); err != nil {
			return chassis.BadRequest(w, "invalid status parameter")
		}
		status = &a
	}
	user := (*string)(nil)
	chassis.StringParam(qs, "user", &user)

	// Normal users can only see their own ownership claims.
	// Administrators see everything and can filter by user ID.
	if user == nil && !authInfo.UserIsAdmin {
		user = &authInfo.UserID
	}
	if user != nil && *user != authInfo.UserID && !authInfo.UserIsAdmin {
		return chassis.Forbidden(w)
	}
	owners := []string{}
	if user != nil {
		usr := *user
		switch usr[0:4] {
		case  "usr_":
			owners, err = s.possibleOwners(usr)
			if err != nil {
				return nil, errors.New("possibleOwners: " + err.Error())
			}
		case "org_":
			owners = append(owners, usr)
		default:
			return nil, errors.New("incorrect format in the owner field")
		}
	}

	claims, total, err := s.db.Claims(owners, status, page, perPage)
	if err != nil {
		return nil, err
	}
	chassis.BuildPaginationResponse(w, r, page, perPage, *total)

	// Make views from claim values.
	result := []*model.ClaimView{}
	ownerIDs := map[string]bool{}
	itemIDs := map[string]bool{}
	for _, claim := range claims {
		result = append(result, model.ViewClaim(&claim, authInfo.UserIsAdmin))
		if authInfo.UserIsAdmin {
			ownerIDs[claim.OwnerID] = true
		}
		itemIDs[claim.ItemID] = true
	}

	// Fill in additional information in claim views.
	var wg sync.WaitGroup
	if len(ownerIDs) > 0 {
		wg.Add(1)
		go s.addOwnerInfo(&wg, result, ownerIDs)
	}
	if len(itemIDs) > 0 {
		wg.Add(1)
		go s.addItemInfo(&wg, result, itemIDs)
	}
	wg.Wait()

	return result, nil
}

// Fill in user information in claim views.
func (s *Server) addOwnerInfo(wg *sync.WaitGroup,
	views []*model.ClaimView, ownerIDs map[string]bool) {
	defer wg.Done()
	ids := []string{}
	for k := range ownerIDs {
		ids = append(ids, k)
	}

	info, err := s.userSvc.Info(ids)
	if err != nil {
		log.Error().Err(err).Msg("failed to fill in owner info for claim views")
	}

	for _, v := range views {
		if i, ok := info[v.ID]; ok {
			if i.Name != nil {
				v.OwnerName = *i.Name
			}
			if i.Email != nil {
				v.OwnerEmail = *i.Email
			}
		}
	}
}

// Fill in item information in claim views.
func (s *Server) addItemInfo(wg *sync.WaitGroup,
	views []*model.ClaimView, itemIDs map[string]bool) {
	defer wg.Done()
	ids := []string{}
	for k := range itemIDs {
		ids = append(ids, k)
	}

	info, err := s.db.ItemNames(ids)
	if err != nil {
		log.Error().Err(err).Msg("failed to fill in item info for claim views")
	}

	for _, v := range views {
		if n, ok := info[v.ItemID]; ok {
			v.ItemName = n
		}
	}
}

// Permissions: administrators can delete any ownership claims; normal
// users can delete only ownership claims that they own.
func (s *Server) deleteOwnershipClaim(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Get authentication information from context and only allow
	// authenticated users to proceed.
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		return chassis.NotFound(w)
	}

	// Get the claim ID to delete.
	claimID := chi.URLParam(r, "id")
	if claimID == "" {
		return chassis.BadRequest(w, "missing claim ID")
	}

	// Determine permission information for update.
	allowedOwners, err := s.allowedOwners(authInfo)
	if err != nil {
		return nil, err
	}

	// Delete the claim.
	err = s.db.DeleteClaim(claimID, allowedOwners)
	if err == db.ErrClaimNotOwned {
		return chassis.Forbidden(w)
	}
	if err == db.ErrClaimNotFound {
		return chassis.NotFound(w)
	}
	if err != nil {
		return nil, err
	}

	return chassis.NoContent(w)
}

type statusChangeRequest struct {
	Status types.ApprovalState `json:"status"`
}

func (s *Server) changeOwnershipClaimStatus(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Get authentication information from context and only allow
	// administrators to proceed.
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod != chassis.SessionAuth || !authInfo.UserIsAdmin {
		return chassis.NotFound(w)
	}

	// Look up claim.
	claimID := chi.URLParam(r, "id")
	claim, err := s.db.ClaimByID(claimID)
	if err == db.ErrClaimNotFound {
		return chassis.NotFound(w)
	}
	if err != nil {
		return nil, err
	}

	// Read and unmarshal request body.
	body, err := chassis.ReadBody(r, 0)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}
	update := statusChangeRequest{}
	err = json.Unmarshal(body, &update)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	// Update claim.
	claim.Status = update.Status
	err = s.db.UpdateClaim(claim)
	if err == db.ErrClaimNotFound {
		return chassis.NotFound(w)
	}
	if err != nil {
		return nil, err
	}

	// If the claim was approved, update the item ownership.
	if update.Status == types.Approved {
		err = s.db.UpdateItemOwnership(claim.ItemID, claim.OwnerID, types.Claimed)
		if err == db.ErrItemNotFound {
			return chassis.NotFound(w)
			log.Error().
				Str("item_id", claim.ItemID).
				Msg("item not found for ownership claim")
		}
		if err != nil {
			return nil, err
		}
	}

	return chassis.NoContent(w)
}
