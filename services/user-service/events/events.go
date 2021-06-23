package events

// Event names used in user service.
const (
	UserCreated          = "user-created"
	UserLogin            = "user-login"
	UserDeleted          = "user-deleted"
	UserUpdated          = "user-updated"
	CreateAPIKey         = "create-api-key"
	DeleteAPIKey         = "delete-api-key"
	OrgUpdated           = "org-updated"
	OrgDeleted           = "org-deleted"
	PayoutAccountCreated = "payout-account-created"
	PayoutAccountUpdated = "payout-account-updated"
	PayoutAccountDeleted = "payout-account-deleted"
	PaymentMethodCreated = "payment-method-created"
	PaymentMethodUpdated = "payment-method-updated"
	PaymentMethodDeleted = "payment-method-deleted"
	CustomerCreated      = "customer-created"
	CustomerDeleted      = "customer-deleted"
	AddressCreated       = "address-created"
	AddressUpdated       = "address-updated"
	AddressDeleted       = "address-deleted"
	DeliveryFeesCreated  = "delivery-fees-created"
	DeliveryFeesUpdated  = "delivery-fees-updated"
	DeliveryFeesDeleted  = "delivery-fees-deleted"
)

// UserCacheInvalTopic is a Pub/Sub topic used to invalidate cached
// user information when users are updated or deleted.
const UserCacheInvalTopic = "invalidate-cached-user"
