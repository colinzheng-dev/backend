package client

import "github.com/veganbase/backend/services/purchase-service/model"

// Client is the service client API for the cart service.
type Client interface {
	UpdatePurchaseStatus(purchaseId string, status string) (*model.Purchase, error)
	UpdateOrderPaymentStatus(orderId string, status string) (*model.Order, error)
	UpdateBookingPaymentStatus(bookingId string, status string) (*model.Booking, error)
	GetPurchaseInfo(purchaseId string) (*model.FullPurchase, error)
	UserBoughtItem(itemId, userId string) (*bool, error)
}

