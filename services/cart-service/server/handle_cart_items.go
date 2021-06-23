package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/cart-service/db"
	"github.com/veganbase/backend/services/cart-service/events"
	"github.com/veganbase/backend/services/cart-service/model"
	itemModel "github.com/veganbase/backend/services/item-service/model"
)

// get all carts items
func (s *Server) cartItemsSearch(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())

	// Get ID from URL parameters.
	cartId := chi.URLParam(r, "cart_id")
	if cartId == "" {
		return chassis.BadRequest(w, "missing cart ID")
	}

	cart, err := s.db.CartByID(cartId)
	if err != nil {
		if err == db.ErrCartNotFound {
			return chassis.NotFoundWithMessage(w, "cart not found")
		}
		return chassis.BadRequest(w, err.Error())
	}

	//if there is no user authenticated, validate if the cart has no owner
	//if there is a session, validate if the user is the owner of the cart
	if (authInfo.AuthMethod == chassis.NoAuth && cart.Owner != "") ||
		(authInfo.AuthMethod == chassis.SessionAuth && authInfo.UserID != cart.Owner) {
		return chassis.BadRequest(w, "user authenticated is not the owner of the cart")
	}

	items, err := s.db.CartItemsByCartId(cart.ID)
	if err != nil {
		return nil, err
	}

	return items, err

}

// Handle normal item search inside a cart.
func (s *Server) cartItemSearch(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())

	// Get ID from URL parameters.
	cartId := chi.URLParam(r, "cart_id")
	cItemId := chi.URLParam(r, "citem_id")

	if cartId == "" {
		return chassis.BadRequest(w, "missing cart ID")
	}
	if cItemId == "" {
		return chassis.BadRequest(w, "missing cart item ID")
	}

	id, err := strconv.Atoi(cItemId)
	if err != nil {
		return chassis.BadRequest(w, "cart item id is invalid. must be an integer")
	}

	//checking if cart exists
	cart, err := s.db.CartByID(cartId)
	if err != nil {
		if err == db.ErrCartNotFound {
			return chassis.NotFoundWithMessage(w, "cart not found")
		}
		return nil, err
	}

	//if there is no user authenticated, validate if the cart has no owner
	//if there is a session, validate if the user is the owner of the cart
	if (authInfo.AuthMethod == chassis.NoAuth && cart.Owner != "") ||
		(authInfo.AuthMethod == chassis.SessionAuth && authInfo.UserID != cart.Owner) {
		return chassis.BadRequest(w, "user authenticated is not the owner of the cart")
	}

	item, err := s.db.CartItemByCartIdAndCartItemId(cartId, id)
	if err != nil {
		return nil, err
	}
	return item, err

}

func (s *Server) addCartItem(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Get authentication information from context
	authInfo := chassis.AuthInfoFromContext(r.Context())

	cartId := chi.URLParam(r, "cart_id")
	if cartId == "" {
		return chassis.BadRequest(w, "missing cart ID")
	}

	// Read request body.
	body, err := chassis.ReadBody(r, 0)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	// Unmarshal request data -- validates JSON request body.
	req := &model.CartItem{}
	err = json.Unmarshal(body, &req)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	if req.ID > 0 {
		return chassis.BadRequest(w, "cart-item ID cannot be manually set")
	}

	//checking if item exists
	itemInfo, err := s.itemSvc.ItemFullWithLink(req.ItemID, "")
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	//checking if the cart exists
	cart, err := s.db.CartByID(cartId)
	if err != nil {
		if err == db.ErrCartNotFound {
			return chassis.NotFoundWithMessage(w, "cart not found")
		}
		return nil, err
	}

	//checking if the cart status allows addition of items
	if cart.CartStatus.String() == "complete" || cart.CartStatus.String() == "abandoned" {
		return chassis.BadRequest(w, "cart status '"+cart.CartStatus.String()+"' does not allow addition of items")
	}

	// unauthenticated carts get auto-claimed
	if cart.Owner == "" && authInfo.UserID != "" {
		cart, err = s.getOrClaimCart(authInfo.UserID, cart)
		if err != nil {
			if err == ErrConvertingCart {
				return chassis.BadRequest(w, err.Error())
			}
			return nil, err
		}
	}

	//if there is no user authenticated, validate if the cart has no owner
	//if there is a session, validate if the user is the owner of the cart
	if (authInfo.AuthMethod == chassis.NoAuth && cart.Owner != "") ||
		(authInfo.AuthMethod == chassis.SessionAuth && authInfo.UserID != cart.Owner) {
		return chassis.BadRequest(w, "user authenticated is not the owner of the cart")
	}

	req.CartID = cart.ID

	//treatment for products( dishes and product-offerings) is different from the other types of items
	//while carts can only have one cart_item for each distinct product-offering, (additions will only increase quantity)
	//every time other type of item is added to the cart, one cart_item will be created
	if req.Type == itemModel.ProductOfferingItem || req.Type == itemModel.DishItem {
		return s.handleProducts(w, req, itemInfo, authInfo.UserID)
	} else {
		return s.handleExperiences(w, req, itemInfo)
	}

}

// handleProductOffering checks if the product is already in the cart. if positive, increase quantity.
func (s *Server) handleProducts(w http.ResponseWriter, req *model.CartItem, itemInfo *itemModel.ItemFullWithLink, user string) (interface{}, error) {
	item, err := s.db.CartItemByCartIdAndItemId(req.CartID, req.ItemID)
	if err != nil {
		if err == db.ErrCartItemNotFound {
			//insert a new item to the cart
			if err = IsAvailableForSale(req.Quantity, *itemInfo); err != nil {
				return chassis.BadRequest(w, "error adding item '"+itemInfo.Name+" to cart :"+err.Error())
			}

			if user != "" {
				if address, err := s.userSvc.GetDefaultAddress(user); err == nil {
					if err = s.CheckDeliveryRegion(address.Coordinates.Latitude, address.Coordinates.Longitude, *itemInfo); err != nil {
						return chassis.BadRequest(w, "error adding item '"+itemInfo.Name+" to cart :"+err.Error())
					}
				} else if err != nil && err.Error() != "default address not found" {
					return chassis.BadRequest(w, "error getting user's default address :"+err.Error())
				}
			}

			if err = s.db.CreateCartItem(req); err != nil {
				return nil, err
			}

			chassis.Emit(s, events.CartItemCreated, req)
			return chassis.NoContent(w)

		}
		return nil, err
	}

	if err = IsAvailableForSale(item.Quantity, *itemInfo); err != nil {
		return chassis.BadRequest(w, "error adding item '"+itemInfo.Name+" to cart :"+err.Error())
	}

	item.Quantity += req.Quantity
	if err = s.db.UpdateCartItem(item); err != nil {
		return nil, err
	}

	chassis.Emit(s, events.CartItemUpdated, item)

	return chassis.NoContent(w)

}

//handleExperiences will add a cart_item without validate if that item is already in the cart
//TODO: in the future may be necessary to check if it is duplicate
func (s *Server) handleExperiences(w http.ResponseWriter, req *model.CartItem, itemInfo *itemModel.ItemFullWithLink) (interface{}, error) {
	var err error

	switch req.Type {
	case itemModel.OfferItem:
		//fixing quantity in case it was omitted at the request
		req.Quantity = 1

		if val, ok := req.OtherInfo["period"]; ok {
			period := val.(map[string]interface{})
			start, end := period["start"].(string), period["end"].(string)
			if err = ValidatePeriod(start, end); err != nil {
				return nil, err
			}
		}
		//checking if start-time is valid
		if _, err = ValidateDatetime(req.OtherInfo["time_start"].(string)); err != nil {
			return nil, err
		}
	case itemModel.RoomItem:
		period := req.OtherInfo["period"].(map[string]interface{})
		start, end := period["start"].(string), period["end"].(string)
		if err = ValidatePeriod(start, end); err != nil {
			return nil, err
		}
		//getting quantity of days based on the difference between end and start date
		if req.Quantity, err = GetNumberOfDays(start, end); err != nil {
			return nil, err
		}
	default:
		return chassis.BadRequest(w, "item_type '"+itemInfo.ItemType.String()+"' is not allowed")
	}

	//checking if item is available
	if err = IsAvailableForSale(req.Quantity, *itemInfo); err != nil {
		return chassis.BadRequest(w, "error adding item '"+itemInfo.Name+" to cart :"+err.Error())
	}
	if err = s.db.CreateCartItem(req); err != nil {
		return nil, err
	}
	chassis.Emit(s, events.CartItemCreated, req)

	return chassis.NoContent(w)

}

// deleteCartItem deletes an item if the user that is logged is the owner of the cart
// or if the cart has no owner.
func (s *Server) deleteCartItem(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Get authentication information from context
	authInfo := chassis.AuthInfoFromContext(r.Context())
	//getting url params
	cartId := chi.URLParam(r, "cart_id")
	cItemId := chi.URLParam(r, "citem_id")
	if cartId == "" {
		return chassis.BadRequest(w, "missing cart ID")
	}
	if cItemId == "" {
		return chassis.BadRequest(w, "missing item ID")
	}

	id, err := strconv.Atoi(cItemId)
	if err != nil {
		return chassis.BadRequest(w, "cart item id is invalid. must be an integer")
	}
	//check if cart exists
	cart, err := s.db.CartByID(cartId)
	if err != nil {
		if err == db.ErrCartNotFound {
			return chassis.NotFoundWithMessage(w, "cart not found")
		}
		return chassis.BadRequest(w, err.Error())
	}

	//check if user is owner of the cart or if the cart is not owned by anybody
	if (authInfo.AuthMethod == chassis.NoAuth && cart.Owner != "") ||
		(authInfo.AuthMethod == chassis.SessionAuth && authInfo.UserID != cart.Owner) {
		return chassis.BadRequest(w, "user authenticated is not the owner of the cart")
	}

	//check if item exists
	item, err := s.db.CartItemByCartIdAndCartItemId(cartId, id)
	if err != nil {
		if err == db.ErrCartItemNotFound {
			return chassis.NotFoundWithMessage(w, "cart item not found")
		}
		return nil, err
	}

	//deleting cart item
	err = s.db.DeleteCartItem(item.ID)
	if err != nil {
		return nil, err
	}
	chassis.Emit(s, events.CartItemDeleted, item)

	return chassis.NoContent(w)
}

// patchCartItem performs a patch operation in a cart item.
// the existence of the cart and item is evaluated, so as the ownership of the cart.
func (s *Server) patchCartItem(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())

	// Get the item ID to update.
	cartId := chi.URLParam(r, "cart_id")
	cItemId := chi.URLParam(r, "citem_id")
	if cartId == "" {
		return chassis.BadRequest(w, "missing cart ID")
	}
	if cItemId == "" {
		return chassis.BadRequest(w, "missing cart item ID")
	}

	id, err := strconv.Atoi(cItemId)
	if err != nil {
		return chassis.BadRequest(w, "cart item id is invalid. must be an integer")
	}

	// Read patch request body.
	body, err := chassis.ReadBody(r, 0)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	// check if the cart exists on the database
	cart, err := s.db.CartByID(cartId)
	if err != nil {
		if err == db.ErrCartNotFound {
			return chassis.NotFoundWithMessage(w, "cart not found")
		}
		return nil, err
	}

	//check if user is owner of the cart or if the cart is not owned by anybody
	if (authInfo.AuthMethod == chassis.NoAuth && cart.Owner != "") ||
		(authInfo.AuthMethod == chassis.SessionAuth && authInfo.UserID != cart.Owner) {
		return chassis.BadRequest(w, "user authenticated is not the owner of the cart")
	}

	item, err := s.db.CartItemByCartIdAndCartItemId(cartId, id)
	if err != nil {
		if err == db.ErrCartItemNotFound {
			return chassis.NotFoundWithMessage(w, "cart item not found")
		}
		return nil, err
	}
	//checking if item exists
	itemInfo, err := s.itemSvc.ItemFullWithLink(item.ItemID, "")
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}
	//patches the item
	if err = item.Patch(body, itemInfo.ItemType); err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	if item.Type == itemModel.OfferItem {
		item.Quantity = 1
		//if period is present, check if start and end dates are valid and if end is after start
		period := item.OtherInfo["period"].(map[string]interface{})

		if len(period) > 0 {
			start, end := period["start"].(string), period["end"].(string)
			if err = ValidatePeriod(start, end); err != nil {
				return nil, err
			}
		}
		//checking if start-time is valid
		if _, err = ValidateDatetime(item.OtherInfo["time-start"].(string)); err != nil {
			return nil, err
		}

	}
	//updating item quantity in case the item is a room
	//we get the quantity based on the number of days between start and end dates
	if item.Type == itemModel.RoomItem {
		period := item.OtherInfo["period"].(map[string]interface{})
		start, end := period["start"].(string), period["end"].(string)
		if err = ValidatePeriod(start, end); err != nil {
			return nil, err
		}
		if item.Quantity, err = GetNumberOfDays(start, end); err != nil {
			return nil, err
		}
	}

	//check if desired quantity is available
	if err = IsAvailableForSale(item.Quantity, *itemInfo); err != nil {
		return chassis.BadRequest(w, "error updating cart item '"+itemInfo.Name+" :"+err.Error())
	}

	if err = s.db.UpdateCartItem(item); err != nil {
		return nil, err
	}

	chassis.Emit(s, events.CartItemUpdated, item)
	return item, nil
}
