package client

import (
	"github.com/veganbase/backend/services/payment-service/model"
	purchase "github.com/veganbase/backend/services/purchase-service/model"
)

// Client is the service client API for the payment service.
type Client interface {
	CreatePaymentIntent(purchase purchase.Purchase) (*[]model.PaymentIntent, error)
}

