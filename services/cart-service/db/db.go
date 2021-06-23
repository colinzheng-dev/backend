package db

import (
	"errors"

	"github.com/veganbase/backend/services/cart-service/model"
)

// ErrCartNotFound is the error returned when an attempt is made to
// access or manipulate a cart with an unknown ID.
var ErrCartNotFound = errors.New("cart ID not found")

var ErrCartItemNotFound = errors.New("cart_item ID not found")

// DB describes the database operations used by the item service.
type DB interface {
	// Retrieval functions for individual carts: by ID or owner 
	CartByID(cartId string) (*model.Cart, error)
	CartsByOwner(owner string) (*[]model.Cart, error)
	GetActiveCartByOwner(owner string) (*model.Cart, error)

	// CreateItem creates a new item.
	CreateCart(cart *model.Cart) (*model.Cart, error)
	// UpdateItem updates the item's details in the database.
	UpdateCart(cart *model.Cart) error
	// DeleteCart deletes a cart in the database if it is not owned by anyone.
	DeleteCart(cartId string) error

	// Retrieval functions for individual cart items: by ID or cartID and itemID
	CartItemByCartIdAndItemId(cartId string, itemId string) (*model.CartItem, error)
	CartItemByCartIdAndCartItemId(cartId string, cItemId int) (*model.CartItem, error)
	CartItemsByCartId(cartId string) (*[]model.CartItem, error)

	// CreateCartItem creates a new cart item
	CreateCartItem(item *model.CartItem) error
	// UpdateCartItem updates a cart item
	UpdateCartItem(item *model.CartItem) error
	// UpdateCartItem updates a cart item.
	DeleteCartItem(id int) error


	// SaveEvent saves an event to the database.
	SaveEvent(topic string, eventData interface{}, inTx func() error) error
}

//go:generate go-bindata -pkg db -o migrations.go migrations/...
