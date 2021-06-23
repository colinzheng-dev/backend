package db

import (
	"errors"
	"github.com/veganbase/backend/chassis"

	"github.com/veganbase/backend/services/purchase-service/model"
)

// ErrPurchaseNotFound is the error returned when an attempt is made to
// access or manipulate a purchase with an unknown ID.
var ErrPurchaseNotFound = errors.New("purchase ID not found")

// ErrOrderNotFound is the error returned when an attempt is made to
// access or manipulate an order with an unknown ID.
var ErrOrderNotFound = errors.New("order ID not found")

// ErrBookingNotFound is the error returned when an attempt is made to
// access or manipulate a booking with an unknown ID.
var ErrBookingNotFound = errors.New("booking ID not found")

// ErrSubscriptionItemNotFound is the error returned when an attempt is made to
// access or manipulate a subscription item with an unknown ID.
var ErrSubscriptionItemNotFound = errors.New("subscription item ID not found")

// ErrSubscriptionPurchaseProcessingNotFound is the error returned when an attempt is made to
// access or manipulate a subscription purchase processing with an unknown ID.
var ErrSubscriptionPurchaseProcessingNotFound = errors.New("subscription purchase processing not found")

type DB interface {
	// PURCHASE
	PurchaseById(purchaseId string) (*model.Purchase, error)
	PurchasesByOwner(ownerId string, params *chassis.Pagination) (*[]model.Purchase, *uint, error)
	CreatePurchase(purchase *model.Purchase, orders *[]model.Order, bookings *[]model.Booking) error
	UpdatePurchase(purchase *model.Purchase) error


	//ORDERS
	OrderById(orderId string) (*model.Order, error)
	OrdersBySeller(sellerId string, params *chassis.Pagination) (*[]model.Order, *uint, error)
	OrdersByPurchase(purchaseId string, params *chassis.Pagination) (*[]model.Order, *uint, error)
	OrdersByBuyer(buyerId string, params *chassis.Pagination) (*[]model.Order, *uint, error)
	UpdateOrder(order *model.Order) error

	// BOOKINGS
	BookingById(bookingId string) (*model.Booking, error)
	BookingsByHost(hostId string, params *chassis.Pagination) (*[]model.Booking, *uint, error)
	BookingsByPurchase(purchaseId string, params *chassis.Pagination) (*[]model.Booking, *uint, error)
	BookingsByBuyer(buyerId string, params *chassis.Pagination) (*[]model.Booking, *uint, error)
	UpdateBooking(booking *model.Booking) error


	// SUBSCRIPTION ITEMS

	SubscriptionItemById(subID string) (*model.SubscriptionItem, error)
	SubscriptionItemsByOwner(ownerID string, params *chassis.Pagination) (*[]model.SubscriptionItem, *uint, error)
	SubscriptionItemsByOwnerAndReference(ownerID string, reference string) (*[]model.SubscriptionItem, error)
	CreateSubscription(sub *model.SubscriptionItem) error
	UpdateSubscriptionItem(sub *model.SubscriptionItem) error
	FlipActivationState(sub *model.SubscriptionItem) error
	DeleteSubscriptionItem(sub *model.SubscriptionItem) error
	UpdateLastDelivery(sub *model.SubscriptionItem) error

	// SUBSCRIPTION PURCHASE
	// processing
	SubscriptionPurchaseProcessingByID(id string) (*model.SubscriptionPurchaseProcessing, error)
	SubscriptionPurchaseProcessingByReference(ref string) (*[]model.SubscriptionPurchaseProcessing, error)
	UpdateSubscriptionPurchaseProcessing(sub *model.SubscriptionPurchaseProcessing) error

	// purchases
	SubscriptionPurchaseByID(purchaseId string) (*model.SubscriptionPurchase, error)
	SubscriptionPurchasesByReferenceAndStatus(ref string, status *string) (*[]model.SubscriptionPurchase, error)
	CreateSubscriptionPurchases(ref string) error
	UpdateSubscriptionPurchase(subs *model.SubscriptionPurchase) error

	// SaveEvent saves an event to the database.
	SaveEvent(topic string, eventData interface{}, inTx func() error) error
}

//go:generate go-bindata -pkg db -o migrations.go migrations/...
