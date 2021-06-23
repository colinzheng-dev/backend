package events

// Event types for purchases, bookings and orders creation and updates.
const (
	PurchaseCreated         = "purchase-created"
	PurchaseUpdated         = "purchase-update"
	OrderCreated            = "order-created"
	OrderUpdated            = "order-update"
	BookingCreated          = "booking-created"
	BookingUpdated          = "booking-update"
	SubscriptionItemUpdated = "subscription-item-updated"
	SubscriptionItemDeleted = "subscription-item-deleted"
)

// Topics that will trigger events on other services.
const (
	PurchaseCreatedTopic = "purchase-created-topic"
	OrderCreatedTopic    = "order-created-topic"
	BookingCreatedTopic  = "booking-created-topic"
	PurchaseUpdatedTopic = "purchase-updated-topic"
)
