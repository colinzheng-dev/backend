package server

import (
	"encoding/json"
	"github.com/go-chi/chi"
	"github.com/rs/zerolog/log"
	"github.com/veganbase/backend/chassis"
	db2 "github.com/veganbase/backend/services/cart-service/db"
	cartUtils "github.com/veganbase/backend/services/cart-service/server"
	it "github.com/veganbase/backend/services/item-service/model"
	"github.com/veganbase/backend/services/purchase-service/db"
	"github.com/veganbase/backend/services/purchase-service/events"
	"github.com/veganbase/backend/services/purchase-service/model"
	"github.com/veganbase/backend/services/purchase-service/model/types"
	usr "github.com/veganbase/backend/services/user-service/model"
	"net/http"
	"net/url"
	"strings"
)

var defaultSite = "https://veganbase.com"
// get all purchases of the logged user
func (s *Server) purchasesSearch(w http.ResponseWriter, r *http.Request) (interface{}, error) {
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
	actionUserID := authInfo.UserID

	purchases, total, err := s.db.PurchasesByOwner(actionUserID, &params)
	if err != nil {
		if err == db.ErrPurchaseNotFound {
			return purchases, nil
		}
		return nil, err
	}

	if qs.Get("format") != "full" {
		chassis.BuildPaginationResponse(w, r, params.Page, params.PerPage, *total)
		return purchases, nil
	}

	var fullView []model.FullPurchase
	for _, pur := range *purchases {
		//getting orders
		var orders *[]model.Order
		if orders, _, err = s.db.OrdersByPurchase(pur.Id, nil); err != nil {
			return nil, err
		}
		//getting bookings
		var bookings *[]model.Booking
		if bookings, _, err = s.db.BookingsByPurchase(pur.Id, nil); err != nil {
			return nil, err
		}
		//adding purchase full view
		fullView = append(fullView, *model.FullView(&pur, orders, bookings, nil))
	}
	chassis.BuildPaginationResponse(w, r, params.Page, params.PerPage, *total)
	return fullView, nil
}

// Handle normal purchase search. This path cannot be accessed without a valid session.
func (s *Server) purchaseSearch(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		return chassis.NotFound(w)
	}
	// Get purchase ID from URL parameters
	purId := chi.URLParam(r, "pur_id")
	if purId == ""{
		return chassis.BadRequest(w, "missing purchase ID")
	}
	//getting purchase
	purchase, err := s.db.PurchaseById(purId)
	if err != nil {
		if err == db.ErrPurchaseNotFound {
			return chassis.NotFoundWithMessage(w, "purchase not found")
		}
		return nil, err
	}
	//only purchases owned by the user can be retrieved
	if authInfo.AuthMethod == chassis.SessionAuth && purchase.BuyerID != authInfo.UserID {
		return chassis.NotFound(w)
	}

	var fullView []model.FullPurchase
	//getting orders
	var orders *[]model.Order
	if orders,_, err = s.db.OrdersByPurchase(purchase.Id, nil); err != nil {
		return nil, err
	}
	//getting bookings
	var bookings *[]model.Booking
	if bookings, _, err = s.db.BookingsByPurchase(purchase.Id, nil); err != nil {
		return nil, err
	}

	fullView = append(fullView, *model.FullView(purchase, orders, bookings, nil))

	return fullView, nil
}

func (s *Server) createPurchase(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Get authentication information from context
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		return chassis.NotFound(w)
	}

	cart, err := s.cartSvc.Active(authInfo.UserID)
	if err != nil {
		if err == db2.ErrCartNotFound {
			return chassis.NotFoundWithMessage(w, "cart not found")
		}
		return chassis.BadRequest(w, "error while getting active cart for the current user: "+ err.Error())
	}
	if len(cart.Items) == 0 {
		return chassis.BadRequest(w, "your cart is empty")
	}

	if !cart.IsValid {
		errors := []string{}
		for _, ci := range cart.Items {
			if ci.HasIssues {
				for _, e := range ci.Errors {
					errors = append(errors, e)
				}
			}
		}
		return chassis.BadRequest(w, "cannot purchase cart with item errors: \n"+ strings.Join(errors, "\n"))
	}

	// Read request body.
	body, err := chassis.ReadBody(r, 0)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	// Unmarshal request data -- validates JSON request body.
	req := &model.PurchaseRequest{}
	if err = json.Unmarshal(body, &req); err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	var address *usr.Address
	if req.AddressId == "" {
		return chassis.BadRequest(w, "you must provide an address")
	} else {
		if address, err = s.userSvc.GetAddress(authInfo.UserID, req.AddressId); err != nil {
			return chassis.BadRequest(w, "address could not be retrieved: " + err.Error())
		}
	}

	var purchaseItems []types.PurchaseItem
	var subscribedItems []model.SubscriptionItem
	itemsByOwner := make(map[string][]types.PurchaseItem) //group items by owner

	deliveriesBySeller := make(map[string]model.DeliveryFee) //group delivery fees by seller
	deliveries := model.DeliveryFees{}

	for _, delivery := range cart.DeliveryFees {
		delivery := model.DeliveryFee{
			Seller:   delivery.Seller,
			Price:    delivery.Price,
			Currency: delivery.Currency,
		}
		deliveries = append(deliveries, delivery)
		deliveriesBySeller[delivery.Seller] = delivery
	}

	for _, cartItem := range cart.Items {
		//checking if item exists
		var itemInfo *it.ItemFullWithLink
		if cartItem.Type == it.ProductOfferingItem {
			if itemInfo, err = s.itemSvc.ItemFullWithLink(cartItem.ItemID, "is-offering-for"); err != nil {
				return chassis.BadRequest(w, "cart item not found: " + cartItem.ItemID)
			}
		} else {
			if itemInfo, err = s.itemSvc.ItemFullWithLink(cartItem.ItemID, ""); err != nil {
				return chassis.BadRequest(w, "cart item not found: " + cartItem.ItemID)
			}
		}
		//creating item entry and adding to the collection
		purInfo := fillPurchaseItem(*itemInfo, cartItem)

		if cartItem.Subscribe {
			subsInfo := fillSubscriptionItem(*itemInfo, cartItem)
			subscribedItems = append(subscribedItems, subsInfo)
		}

		purchaseItems = append(purchaseItems, purInfo)

		//if the purchase item is part of an order, group it by owner
		//this may consume more memory, but will be faster to create orders and ease the marshaling process
		if purInfo.ProductType == it.ProductOfferingItem.String() || purInfo.ProductType == it.DishItem.String() {
			itemsByOwner[itemInfo.Owner.ID] = append(itemsByOwner[itemInfo.Owner.ID], purInfo)
		}
	}
	//creating purchase
	purchase := model.Purchase{}
	purchase.BuyerID = cart.Owner
	err = purchase.Status.FromString("pending")
	purchase.Items = purchaseItems
	purchase.DeliveryFees = &deliveries
	purchase.Site = &defaultSite
	if origins, ok := r.Header["Origin"]; ok && len(origins) > 0 {
		purchase.Site = &origins[0]
	}

	//creating orders and bookings
	orders, bookings := fillOrdersAndBookings(purchase.BuyerID, itemsByOwner, deliveriesBySeller, purchaseItems, address)

	if err = s.db.CreatePurchase(&purchase, &orders, &bookings); err != nil {
		return nil, err
	}

	// emitting events
	chassis.Emit(s, events.PurchaseCreated, purchase)
	for _, o := range orders {
		chassis.Emit(s, events.OrderCreated, o)
	}
	for _, b := range bookings {
		chassis.Emit(s, events.BookingCreated, b)
	}

	//updating stock availability
	for _, cartItem := range cart.Items {
		if err = s.itemSvc.UpdateItemAvailability(cartItem.ItemID, -cartItem.Quantity); err != nil {
			//TODO: do nothing or log
		}
	}

	//finishing cart
	s.cartSvc.FinishCart(cart.ID)

	//trigger payments
	intents, _ := s.paymentSvc.CreatePaymentIntent(purchase)

	// TODO: handle payment errors. not sure if we must break the flow here or just log somewhere to enable
	// the user to fixing payment later

	if intents != nil {
		paid := true
		//checking if all payments are successful
		for _, intent := range *intents {
			if intent.Status != "success" {
				paid = false
			}
		}

		//if true, set purchase status to 'completed'
		if paid == true {
			if err = purchase.Status.FromString("completed"); err != nil {
				return nil, err
			}
			if err = s.db.UpdatePurchase(&purchase); err != nil {
				return nil, err
			}
		}
	}

	//return complete purchase view
	purchaseView := model.FullView(&purchase, &orders, &bookings, intents)

	//this routine will trigger email notifications separately
	go s.sendPurchaseCreatedNotifications(purchaseView)
	//this routine will add subscription items to the user's subscription list
	go s.addSubscriptionItems(subscribedItems, purchase.Id, address.ID)
	return purchaseView, err

}

func (s *Server) addSubscriptionItems(items []model.SubscriptionItem, purchaseID, addressID string) {
	for _, i := range items {
		i.Origin = purchaseID
		i.AddressID = &addressID
		if err := s.db.CreateSubscription(&i);  err != nil {
			log.Error().Err(err).Msg("add-subscription-item: error while creating subscription item")
		}
	}
}

func (s *Server) createSimplePurchase(w http.ResponseWriter, r *http.Request) (interface{}, error) {

	// Read request body.
	body, err := chassis.ReadBody(r, 0)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	// Unmarshal request data -- validates JSON request body.
	req := &model.SimplePurchaseRequest{}
	if err = json.Unmarshal(body, &req); err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	//getting user information or creating a new one with default values
	rsp, err := s.userSvc.Login(req.Email, "","")
	if err != nil {
		return nil, err
	}

	//checking if item exists
	itemInfo, err := s.itemSvc.ItemFullWithLink(req.ItemID, "")
	if err != nil {
		return chassis.BadRequest(w, "item not found: " + req.ItemID)
	}
	//check if item is a subtype of products.
	//their sale depends on the address, therefore cannot be made here.
	if itemInfo.ItemType == it.ProductOfferingItem || itemInfo.ItemType == it.DishItem {
		return chassis.BadRequest(w, "cannot purchase " + itemInfo.ItemType.String() + " using this method")
	}

	//checking if quantity is available in stock
	if err = cartUtils.IsAvailableForSale(req.Quantity, *itemInfo); err != nil {
		return chassis.BadRequest(w, "error while processing item '"+itemInfo.Name+" :" + err.Error())
	}

	//creating purchase item entry and adding to the collection
	purInfo := types.PurchaseItem{
		ItemId:      itemInfo.ID,
		ProductType: itemInfo.ItemType.String(),
		ItemOwner:   itemInfo.Owner.ID,
		Quantity:    req.Quantity,
		Price:       int(itemInfo.Attrs["price"].(float64)),
		Currency:    itemInfo.Attrs["currency"].(string),
		OtherInfo:   make(map[string]interface{}),
	}
	if req.OtherInfo != nil {
		for k, v := range req.OtherInfo {
			purInfo.OtherInfo[k] = v
		}
	}

	//creating purchase
	purchase := model.Purchase{}
	purchase.Items = types.PurchaseItems{purInfo}
	purchase.BuyerID = rsp.User.ID
	purchase.Status.FromString("pending")
	purchase.Site = &defaultSite
	purchase.PaymentMethod = req.PaymentMethodID

	if origins, ok := r.Header["Origin"]; ok && len(origins) > 0 {
		purchase.Site = &origins[0]
	}

	orders, bookings := fillOrdersAndBookings(rsp.ID, map[string][]types.PurchaseItem{},nil, purchase.Items, nil )
	if err = s.db.CreatePurchase(&purchase, &orders, &bookings); err != nil {
		return nil, err
	}

	// emitting events
	chassis.Emit(s, events.PurchaseCreated, purchase)
	for _, o := range orders {
		chassis.Emit(s, events.OrderCreated, o)
	}
	for _, b := range bookings {
		chassis.Emit(s, events.BookingCreated, b)
	}

	//updating stock availability
	if err = s.itemSvc.UpdateItemAvailability(req.ItemID, - req.Quantity); err != nil {
		//TODO: do nothing or log
	}

	//trigger payments
	intents, _ := s.paymentSvc.CreatePaymentIntent(purchase)
	// TODO: handle payment errors. not sure if we must break the flow here or just log somewhere to enable
	//  the user to fixing payment later

	if intents != nil {
		paid := true
		//checking if all payments are successful
		for _, intent := range *intents {
			if intent.Status != "success" {
				paid = false
			}
		}

		//if true, set purchase status to 'completed'
		if paid == true {
			if err = purchase.Status.FromString("completed"); err != nil {
				return nil, err
			}
			if err = s.db.UpdatePurchase(&purchase); err != nil {
				return nil, err
			}
		}
	}

	//return complete purchase view
	purchaseView := model.FullView(&purchase, &orders, &bookings, intents)

	//this routine will trigger email notifications separately
	go s.sendPurchaseCreatedNotifications(purchaseView)

	return purchaseView, err
}

func (s *Server) patchPurchase(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		return chassis.NotFound(w)
	}

	// Get the item ID to update.
	purId := chi.URLParam(r, "pur_id")
	// Read patch request body.
	body, err := chassis.ReadBody(r, 0)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	// check if the purchase exists on the database
	purchase, err := s.db.PurchaseById(purId)
	if err != nil {
		if err == db.ErrPurchaseNotFound {
			return chassis.NotFoundWithMessage(w, "purchase not found")
		}
		return nil, err
	}

	if err = purchase.Patch(body); err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	//Do the update.
	if err = s.db.UpdatePurchase(purchase); err != nil {
		return nil, err
	}

	chassis.Emit(s, events.PurchaseUpdated, purchase)

	return purchase, nil
}


func (s *Server) patchPurchaseInternal(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Get the item ID to update.
	purId := chi.URLParam(r, "pur_id")
	if purId == "" {
		return chassis.BadRequest(w, "purchase id is missing")
	}
	// Read patch request body.
	body, err := chassis.ReadBody(r, 0)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	// check if the purchase exists on the database
	purchase, err := s.db.PurchaseById(purId)
	if err != nil {
		if err == db.ErrPurchaseNotFound {
			return chassis.NotFoundWithMessage(w, "purchase not found")
		}
		return nil, err
	}


	if err = purchase.Patch(body); err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	//Do the update.
	if err = s.db.UpdatePurchase(purchase); err != nil {
		return nil, err
	}

	chassis.Emit(s, events.PurchaseUpdated, purchase)

	return purchase, nil
}
//purchaseSearchInternal is the same as the public one, except for the auth and ownership validations.
func (s *Server) purchaseSearchInternal(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Get purchase ID from URL parameters
	purId := chi.URLParam(r, "pur_id")
	qs, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return chassis.BadRequest(w, "error while parsing the query")
	}
	//getting purchase
	purchase, err := s.db.PurchaseById(purId)
	if err != nil {
		if err == db.ErrPurchaseNotFound {
			return chassis.NotFoundWithMessage(w, "purchase not found")
		}
		return nil, err
	}

	if qs.Get("format") != "full" {
		return purchase, nil
	}

	//getting orders
	var orders *[]model.Order
	if orders, _, err = s.db.OrdersByPurchase(purchase.Id, nil); err != nil {
		return nil, err
	}
	//getting bookings
	var bookings *[]model.Booking
	if bookings,_, err = s.db.BookingsByPurchase(purchase.Id, nil); err != nil {
		return nil, err
	}

	return *model.FullView(purchase, orders, bookings, nil), nil
}