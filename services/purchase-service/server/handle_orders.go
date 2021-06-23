package server

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog/log"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/purchase-service/db"
	"github.com/veganbase/backend/services/purchase-service/events"
	"github.com/veganbase/backend/services/purchase-service/model"
	"github.com/veganbase/backend/services/purchase-service/model/types"
)

// get all orders of a specific purchase
func (s *Server) ordersByPurchaseSearch(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		return chassis.NotFound(w)
	}

	purId := chi.URLParam(r, "pur_id")
	if purId == "" {
		return chassis.BadRequest(w, "missing purchase ID")
	}

	qs, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return chassis.BadRequest(w, "error while parsing the query")
	}

	params := chassis.Pagination{}
	// Pagination parameters.
	if err := chassis.PaginationParams(qs, &params.Page, &params.PerPage); err != nil {
		return nil, err
	}

	purchase, err := s.db.PurchaseById(purId)
	if err != nil {
		if err == db.ErrPurchaseNotFound {
			return chassis.NotFoundWithMessage(w, "purchase not found")
		}
		return nil, err
	}

	if authInfo.UserID != purchase.BuyerID {
		return chassis.NotFound(w)
	}

	orders, total, err := s.db.OrdersByPurchase(purchase.Id, &params)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}
	chassis.BuildPaginationResponse(w, r, params.Page, params.PerPage, *total)
	return s.BuildOrderFullResponse(orders)

}

func (s *Server) ordersSearch(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		return chassis.NotFound(w)
	}

	qs, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return chassis.BadRequest(w, "error while parsing the query")
	}

	params := chassis.Pagination{}

	// Pagination parameters.
	if err := chassis.PaginationParams(qs, &params.Page, &params.PerPage); err != nil {
		return nil, err
	}
	seller := qs.Get("seller")
	org := qs.Get("org")

	//getting all orders that the logged in user is the seller
	if seller == "true" {
		owner := authInfo.UserID
		if org != "" {
			isOrgMember, err := s.userSvc.IsUserOrgMember(authInfo.UserID, org)
			if err != nil {
				return chassis.BadRequest(w, err.Error())
			}
			if !isOrgMember {
				return chassis.BadRequest(w, "user not member of organisation")
			}
			owner = org
		}
		orders, total, err := s.db.OrdersBySeller(owner, &params)
		if err != nil {
			return chassis.BadRequest(w, err.Error())
		}
		chassis.BuildPaginationResponse(w, r, params.Page, params.PerPage, *total)
		return s.BuildOrderFullResponse(orders)
	}

	//getting all orders of the current user
	orders, total, err := s.db.OrdersByBuyer(authInfo.UserID, &params)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}
	chassis.BuildPaginationResponse(w, r, params.Page, params.PerPage, *total)
	return s.BuildOrderFullResponse(orders)
}

func (s *Server) canViewOrder(order *model.Order, userID string) (bool, error) {
	if userID == order.BuyerID {
		return true, nil
	}

	orgMember, err := s.userSvc.IsUserOrgMember(userID, order.Seller)
	if err != nil {
		return false, err
	}

	if orgMember {
		return true, nil
	}

	return false, nil
}

// Handle normal order search.
func (s *Server) orderSearch(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		return chassis.NotFound(w)
	}

	ordId := chi.URLParam(r, "ord_id")
	if ordId == "" {
		return chassis.BadRequest(w, "missing order ID")
	}
	//checking if order exists
	order, err := s.db.OrderById(ordId)
	if err != nil {
		if err == db.ErrOrderNotFound {
			return chassis.NotFoundWithMessage(w, "order not found")
		}
		return chassis.BadRequest(w, err.Error())
	}

	canView, err := s.canViewOrder(order, authInfo.UserID)
	if err != nil {
		return chassis.InternalServerError(w, err)
	}

	if !canView {
		log.Info().Msgf("user %s is not authorized to view order %s", authInfo.UserID, order.Id)
		return chassis.NotFound(w)
	}

	//calling user service to get updated information about users
	userInfo, err := s.userSvc.Info([]string{order.BuyerID, order.Seller})
	if err != nil {
		return order, nil
	}

	itemIDs := []string{}
	for _, it := range order.Items {
		itemIDs = append(itemIDs, it.ItemId)
	}
	itemInfo, err := s.itemSvc.GetItemsInfo(itemIDs)
	if err != nil {
		return order, nil
	}

	fullItems := types.FullPurchaseItems{}
	for _, it := range order.Items {
		fullItems = append(fullItems, *types.GetFullPurchaseItem(&it, itemInfo[it.ItemId]))
	}
	return *model.GetFullOrder(order, userInfo[order.BuyerID], userInfo[order.Seller], fullItems), err

}

// patchOrder performs a patch operation in an order
// the existence of the order is evaluated, so as its ownership and the link with purchase.
func (s *Server) patchOrder(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		return chassis.NotFound(w)
	}

	ordId := chi.URLParam(r, "ord_id")

	// Read patch request body.
	body, err := chassis.ReadBody(r, 0)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	// check if the order exists on the database
	order, err := s.db.OrderById(ordId)
	if err != nil {
		if err == db.ErrOrderNotFound {
			return chassis.NotFoundWithMessage(w, "order not found")
		}
		return nil, err
	}

	//check if user is owner of the cart or if the cart is not owned by anybody
	if authInfo.UserID != order.BuyerID {
		return chassis.BadRequest(w, "user authenticated is not the owner of the order")
	}

	err = order.Patch(body)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	// Do the update.
	if err = s.db.UpdateOrder(order); err != nil {
		return nil, err
	}

	chassis.Emit(s, events.OrderUpdated, order)

	return order, nil
}

// patchOrderInternal performs a patch operation in an order
// the existence of the order is evaluated, so as its ownership and the link with purchase.
func (s *Server) patchOrderInternal(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	ordId := chi.URLParam(r, "ord_id")

	// Read patch request body.
	body, err := chassis.ReadBody(r, 0)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	// check if the order exists on the database
	order, err := s.db.OrderById(ordId)
	if err != nil {
		if err == db.ErrOrderNotFound {
			return chassis.NotFoundWithMessage(w, "order not found")
		}
		return nil, err
	}

	err = order.Patch(body)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	// Do the update.
	if err = s.db.UpdateOrder(order); err != nil {
		return nil, err
	}

	chassis.Emit(s, events.OrderUpdated, order)

	return order, nil
}

func (s *Server) BuildOrderFullResponse(rawOrders *[]model.Order) (interface{}, error) {
	if len(*rawOrders) == 0 {
		return rawOrders, nil
	}
	//Getting unique ids to acquire full user information
	userIDs := map[string]bool{}
	itemIDs := map[string]bool{}
	for _, ord := range *rawOrders {
		userIDs[ord.BuyerID] = true
		userIDs[ord.Seller] = true
		for _, it := range ord.Items {
			itemIDs[it.ItemId] = true
		}
	}

	uniqueUserIDs := []string{}
	uniqueItemIDs := []string{}
	for k := range userIDs {
		uniqueUserIDs = append(uniqueUserIDs, k)
	}

	for k := range itemIDs {
		uniqueItemIDs = append(uniqueItemIDs, k)
	}
	//calling user service to get updated information about users
	userInfo, err := s.userSvc.Info(uniqueUserIDs)
	if err != nil {
		return rawOrders, nil
	}

	itemInfo, err := s.itemSvc.GetItemsInfo(uniqueItemIDs)
	if err != nil {
		return rawOrders, nil
	}

	var fullOrders []model.FullOrder
	for _, ord := range *rawOrders {
		fullItems := types.FullPurchaseItems{}
		for _, it := range ord.Items {
			ii := itemInfo[it.ItemId]
			if ii == nil {
				log.Warn().Msg(fmt.Sprintf("item %s is in order %s but not returned by itemSvc", it.ItemId, ord.Id))
				continue
			}
			fullItems = append(fullItems, *types.GetFullPurchaseItem(&it, ii))
		}
		fullOrders = append(fullOrders, *model.GetFullOrder(&ord, userInfo[ord.BuyerID], userInfo[ord.Seller], fullItems))
	}

	return fullOrders, nil

}
