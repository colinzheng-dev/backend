package client

import "github.com/veganbase/backend/services/cart-service/model"

// Client is the service client API for the cart service.
type Client interface {
	Active(userId string) (*model.FullCart, error)
	FinishCart(cartId string) (*model.Cart, error)
}

