package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/veganbase/backend/services/cart-service/db"
	"github.com/veganbase/backend/services/cart-service/events"
	"github.com/veganbase/backend/services/cart-service/model/types"

	"github.com/go-chi/chi"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/cart-service/model"
)

// get all carts of the logged user, without their items
func (s *Server) cartsSearch(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		return chassis.NotFound(w)
	}
	carts, err := s.db.CartsByOwner(authInfo.UserID)
	if err != nil {
		return nil, err
	}

	return carts, nil
}

// Handle normal cart search. This path can be accessed without a valid session, but only
// carts with no owner can be retrieved.
func (s *Server) cartSearch(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())

	// Get ID from URL parameters.
	cartId := chi.URLParam(r, "cart_id")

	cart, err := s.db.CartByID(cartId)
	if err != nil {
		if err == db.ErrCartNotFound {
			return chassis.NotFoundWithMessage(w, "cart not found")
		}
		return nil, err
	}

	items, err := s.db.CartItemsByCartId(cartId)
	if err != nil {
		return nil, err
	}
	var errs *map[int][]string
	var fees *[]model.DeliveryFee
	if cart.CartStatus == types.NotLoggedIn || cart.CartStatus == types.Active {
		errs, err = s.CheckInvalidItems(authInfo.UserID, *items)
		if err != nil {
			return nil, err
		}

		fees, err = s.CalculateDeliveryFees(*items)
		if err != nil {
			return nil, err
		}
	}
	view := model.FullView(cart, items, fees, *errs)

	//if the user is not logged in, only "not logged in" carts can be retrieved
	//if the user is logged in, only carts owned by him can be retrieved
	//TODO: currently not logged in users can retrieve other carts with no owner
	if (authInfo.AuthMethod == chassis.NoAuth && cart.Owner != "") ||
		(authInfo.AuthMethod == chassis.SessionAuth && cart.Owner != authInfo.UserID) {
		return chassis.BadRequest(w, "the user is not the owner of the cart")
	}
	return view, nil
}

func (s *Server) getActiveCart(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		return chassis.NotFound(w)
	}

	//check if user has an active cart
	activeCart, err := s.db.GetActiveCartByOwner(authInfo.UserID)
	if err != nil {
		if err == db.ErrCartNotFound {
			return chassis.NotFoundWithMessage(w, "the user has no active cart")
		}
		return nil, err
	}

	items, err := s.db.CartItemsByCartId(activeCart.ID)
	if err != nil {
		return nil, err
	}

	errs, err := s.CheckInvalidItems(authInfo.UserID, *items)
	if err != nil {
		return nil, err
	}

	fees, err := s.CalculateDeliveryFees(*items)
	if err != nil {
		return nil, err
	}

	view := model.FullView(activeCart, items, fees, *errs)

	return view, err
}

func (s *Server) createCart(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Get authentication information from context
	authInfo := chassis.AuthInfoFromContext(r.Context())

	req := &model.Cart{}
	// Read request body.
	if body, err := chassis.ReadBody(r, 0); body != nil {
		// Unmarshal request data -- validates JSON request body.
		err = json.Unmarshal(body, &req)
		if err != nil {
			return chassis.BadRequest(w, err.Error())
		}
	}

	//creating a cart without an user logged in
	//the cart_id must be stored on a cookie
	if authInfo.AuthMethod == chassis.NoAuth {
		req.Owner = ""
		err := req.CartStatus.FromString("not logged in")
		if err != nil {
			return nil, err
		}

		cart, err := s.db.CreateCart(req)
		if err != nil {
			return nil, err
		}
		chassis.Emit(s, events.CartCreated, req)
		return cart, nil

	} else if authInfo.AuthMethod == chassis.SessionAuth {
		active, err := s.db.GetActiveCartByOwner(authInfo.UserID)
		if err != nil {
			if err == db.ErrCartNotFound {
				err = req.CartStatus.FromString("active")
				if err != nil {
					return nil, err
				}

				req.Owner = authInfo.UserID
				cart, err := s.db.CreateCart(req)
				if err != nil {
					return nil, err
				}
				chassis.Emit(s, events.CartCreated, req)
				return cart, nil
			}
			return nil, err
		}
		if active != nil {
			return active, nil
		}
		return nil, err

	}
	return chassis.BadRequest(w, "auth method not valid")
}

func (s *Server) forgetCart(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		return chassis.NotFound(w)
	}

	//check if user has an active cart
	activeCart, err := s.db.GetActiveCartByOwner(authInfo.UserID)
	if err != nil {
		if err == db.ErrCartNotFound {
			return chassis.BadRequest(w, "the user has no active cart to forget")
		}
		return nil, err
	}

	if err = activeCart.CartStatus.FromString("abandoned"); err != nil {
		return nil, err
	}
	//update cart status
	if err = s.db.UpdateCart(activeCart); err != nil {
		return nil, err
	}
	chassis.Emit(s, events.CartUpdated, activeCart)
	return chassis.NoContent(w)
}

func (s *Server) mergeCarts(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		return chassis.NotFound(w)
	}

	// Get the item ID to update.
	tempCartId := chi.URLParam(r, "cart_id")
	if tempCartId == "" {
		return chassis.BadRequest(w, "missing anonymous cart ID")
	}

	// Look up item value and patch it.
	tempCart, err := s.db.CartByID(tempCartId)
	if err != nil {
		if err == db.ErrCartNotFound {
			return chassis.NotFoundWithMessage(w, "anonymous cart not found")
		}
		return nil, err
	}

	if tempCart.Owner != "" && tempCart.Owner != authInfo.UserID {
		return chassis.BadRequest(w, "cannot merge a cart that already has an owner")
	}

	activeCart, err := s.getOrClaimCart(authInfo.UserID, tempCart)
	if err != nil {
		if err == ErrConvertingCart {
			return chassis.BadRequest(w, err.Error())
		}
		return nil, err
	}

	if activeCart != nil && activeCart.ID != tempCart.ID {

		itemsOnTempCart, err := s.db.CartItemsByCartId(tempCart.ID)
		if err != nil {
			return nil, err
		}

		//get items on active cart
		itemsOnActiveCart, err := s.db.CartItemsByCartId(activeCart.ID)
		if err != nil {
			return nil, err
		}
		//TODO: CHANGE TO DO THINS IN TRANSACTION
		for _, tempItem := range *itemsOnTempCart {
			isInBothCarts := false
			for _, item := range *itemsOnActiveCart {
				//if item is in both carts, update the quantity on active cart
				if tempItem.ItemID == item.ItemID {
					isInBothCarts = true
					item.Quantity += tempItem.Quantity
					_ = s.db.UpdateCartItem(&item)
					continue
				}
			}
			if !isInBothCarts {
				tempItem.CartID = activeCart.ID
				_ = s.db.UpdateCartItem(&tempItem)
			}
		}
		chassis.Emit(s, events.CartMerged, tempCart)
		_ = s.db.DeleteCart(tempCart.ID)
	}

	items, err := s.db.CartItemsByCartId(activeCart.ID)
	if err != nil {
		return nil, err
	}

	errs, err := s.CheckInvalidItems(authInfo.UserID, *items)
	if err != nil {
		return nil, err
	}

	fees, err := s.CalculateDeliveryFees(*items)
	if err != nil {
		return nil, err
	}
	return model.FullView(activeCart, items, fees, *errs), nil
}

func (s *Server) patchCart(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())

	// Get the item ID to update.
	cartId := chi.URLParam(r, "cart_id")
	if cartId == "" {
		return chassis.BadRequest(w, "missing cart ID")
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

	if err = cart.Patch(body); err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	// Do the update.
	if err = s.db.UpdateCart(cart); err != nil {
		return nil, err
	}

	chassis.Emit(s, events.CartUpdated, cart)

	return cart, nil
}

// internalPatchCart is an internal usage only that patches a cart without session validations
func (s *Server) internalPatchCart(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Get the item ID to update.
	cartId := chi.URLParam(r, "cart_id")
	if cartId == "" {
		return chassis.BadRequest(w, "missing cart ID")
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

	if err = cart.Patch(body); err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	// Do the update.
	if err = s.db.UpdateCart(cart); err != nil {
		return nil, err
	}

	chassis.Emit(s, events.CartUpdated, cart)

	return cart, nil
}

// internalActive is an unprotected path to be called by the purchase-service.
// It will return the active cart of a certain user
func (s *Server) internalActive(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	userId := chi.URLParam(r, "user_id")
	//check if user has an active cart
	activeCart, err := s.db.GetActiveCartByOwner(userId)
	if err != nil {
		if err == db.ErrCartNotFound {
			return chassis.NotFoundWithMessage(w, "the user has no active cart")
		}
		return nil, err
	}

	items, err := s.db.CartItemsByCartId(activeCart.ID)
	if err != nil {
		return nil, err
	}

	errs, err := s.CheckInvalidItems(userId, *items)
	if err != nil {
		return nil, err
	}

	fees, err := s.CalculateDeliveryFees(*items)
	if err != nil {
		return nil, err
	}

	return model.FullView(activeCart, items, fees, *errs), err
}

var ErrConvertingCart = fmt.Errorf("error converting cart status")

func (s *Server) getOrClaimCart(userID string, tempCart *model.Cart) (*model.Cart, error) {
	activeCart, err := s.db.GetActiveCartByOwner(userID)
	if err != nil {
		if err == db.ErrCartNotFound && (tempCart.Owner == "" || tempCart.Owner == userID) {
			//the user doesn't have an active cart. setting temp cart as valid active cart
			tempCart.Owner = userID

			if err := tempCart.CartStatus.FromString("active"); err != nil {
				return nil, ErrConvertingCart
			}

			if err := s.db.UpdateCart(tempCart); err != nil {
				return nil, err
			}
			chassis.Emit(s, events.CartUpdated, tempCart)
			return tempCart, nil
		}
		return nil, err
	}

	return activeCart, nil
}
