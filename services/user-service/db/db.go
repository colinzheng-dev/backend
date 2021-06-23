package db

import (
	"github.com/pkg/errors"
	"github.com/veganbase/backend/services/user-service/messages"
	"github.com/veganbase/backend/services/user-service/model"
)

// ErrUserNotFound is the error returned when an attempt is made to
// access or manipulate a user with an unknown ID.
var ErrUserNotFound = errors.New("user ID not found")

// ErrOrgNotFound is the error returned when an attempt is made to
// access or manipulate an organisation with an unknown ID.
var ErrOrgNotFound = errors.New("organisation not found")

// ErrPayoutAccountNotFound is the error returned when an attempt is made to
// access or manipulate a payout account with an unknown ID.
var ErrPayoutAccountNotFound = errors.New("payout account not found")

// ErrPaymentMethodNotFound is the error returned when an attempt is made to
// access or manipulate a payment method with an unknown ID.
var ErrPaymentMethodNotFound = errors.New("payout method not found")

// ErrCustomerNotFound is the error returned when an attempt is made to
// access or manipulate a nonexistent customer.
var ErrCustomerNotFound = errors.New("customer not found")

// ErrAddressNotFound is the error returned when an attempt is made to
// access or manipulate an address with an unknown ID.
var ErrAddressNotFound = errors.New("address not found")


// ErrDeliveryFeesNotFound is the error returned when an attempt is made to
// access or manipulate an delivery fee with an unknown ID or owner.
var ErrDeliveryFeesNotFound = errors.New("delivery fees not found")

// ErrReadOnlyField is the error returned when an attempt is made to
// update a read-only field for a user (e.g. email, last login time,
// API key).
var ErrReadOnlyField = errors.New("attempt to modify read-only field")

// ErrUserAlreadyInOrg is the error returned when an attempt is made
// to add a user to an organisation when the user is already a member
// of the organisation.
var ErrUserAlreadyInOrg = errors.New("user is already a member of organisation")

// DB describes the database operations used by the user service.
type DB interface {
	// UserByID returns the full user model for a given user ID.
	UserByID(id string) (*model.User, error)

	// UsersByIDs returns the full user model for a given list of user
	// IDs as a map indexed by the ID.
	UsersByIDs(ids []string) (map[string]*model.User, error)

	// UserByAPIKey returns a user with an specific apy_key
	UserByAPIKey(apiKey string) (*model.User, error)

	// Users gets a list of users in reverse last login date order,
	// optionally filtered by a search term and paginated.
	Users(search string, page, perPage uint) ([]model.User, error)

	// Info gets minimal information about a list of users and
	// organisations given their IDs.
	Info(ids []string) (map[string]model.Info, error)

	// LoginUser performs login actions for a given email address:
	//
	//  - If an account with the given email address already exists in the
	//    database, then set the account's last_login field to the current time.
	//
	//  - If an account with the given email address does not already
	//    exist in the database, then create a new user account with the
	//    given email address, defaulting the name and display_name
	//    fields to match the email address and setting the new
	//    account's last_login field to the current time.
	//
	// In both cases, return the full user record of the logged in user.
	LoginUser(email string, avatarGen func() string) (*model.User, bool, error)

	// UpdateUser updates the user's details in the database. The id,
	// email, last_login and api_key fields are read-only using this
	// method.
	UpdateUser(user *model.User) error

	// DeleteUser deletes the given user account.
	// TODO: ADD SOME SORT OF ARCHIVAL MECHANISM INSTEAD.
	DeleteUser(id string) error

	SaveHashedAPIKey(key, secret, userID string) error
	DeleteAPIKey(id string) error

	// CreateOrg creates a new organisation.
	CreateOrg(org *model.Organisation) error

	// UpdateOrg updates an organisations's details in the database. The
	// id, slug, approval and created_at fields are read-only using this
	// method.
	UpdateOrg(org *model.Organisation) error

	// DeleteOrg deletes an organisation.
	DeleteOrg(orgID string) error

	// OrgByID returns the full organisation model for a given
	// organisation ID.
	//OrgByID(id string) (*model.Organisation, error)


	// OrgByIDorSlug returns the full organisation model for a given
	// organisation ID.
	OrgByIDorSlug(idOrSlug string) (*model.Organisation, error)

	// Orgs gets a list of organisations in alphabetical name order,
	// optionally filtered by a search term and paginated.
	Orgs(search string, page, perPage uint) ([]*model.Organisation, error)

	// OrgUsers gets a list of all users in an organisation.
	OrgUsers(id string) ([]*model.OrgUser, error)

	// UserOrgs gets a list of organisations of which a user is a
	// member.
	UserOrgs(userID string) ([]*model.OrgWithUserInfo, error)

	// OrgAddUser adds a user to an organisation.
	OrgAddUser(orgID string, userID string, isOrgAdmin bool) error

	// OrgPatchUser updates a user's admin status within an organisation.
	OrgPatchUser(orgID string, userID string, isOrgAdmin bool) error

	// OrgDeleteUser removes a user from an organisation.
	OrgDeleteUser(orgID string, userID string) error

	RotateSSOSecret(secret, orgID string) error
	DeleteSSOSecret(id string) error
	GetSSOSecretByOrgIDOrSlug(id string) (*messages.SSOSecret, error)

	//Payout-accounts
	CreatePayoutAccount(acc *model.PayoutAccount) error
	PayoutAccountByOwner(ownerId string) (*model.PayoutAccount, error)
	PayoutAccountById(id string) (*model.PayoutAccount, error)
	UpdatePayoutAccount(acc *model.PayoutAccount) error
	DeletePayoutAccount(id string) error

	//Payment-methods
	CreatePaymentMethod(pmt *model.PaymentMethod) error
	PaymentMethodsByUserId(id string) (*[]model.PaymentMethod, error)
	DefaultPaymentMethodByUserId(id string) (*model.PaymentMethod, error)
	PaymentMethodById(id string) (*model.PaymentMethod, error)
	UpdatePaymentMethod(pmt *model.PaymentMethod) error
	DeletePaymentMethod(id string) error

	//Customers
	CustomerByUserId(userId string) (*model.Customer, error)
	CreateCustomer(cus *model.Customer) error
	DeleteCustomer(userId string) error

	//Addresses
	CreateAddress(addr *model.Address) error
	AddressById(id string) (*model.Address, error)
	DefaultAddressByUserId(id string) (*model.Address, error)
	AddressesByUserId(id string) (*[]model.Address, error)
	UpdateAddress(addr *model.Address) error
	DeleteAddress(id string) error
	NotificationInfoByUserId(userId string) (*model.EmailNotificationInfo, error)
	NotificationInfoByOrgId(orgId string) (*model.EmailNotificationInfo, error)

	//Delivery fees
	CreateDeliveryFees(delFee *model.DeliveryFees) error
	DeliveryFeesByOwner(id string) (*model.DeliveryFees, error)
	DeliveryFeesById(id string) (*model.DeliveryFees, error)
	UpdateDeliveryFees(delFee *model.DeliveryFees) error
	DeleteDeliveryFees(id string) error
	//used internally to get delivery fees of multiple users/orgs with one request
	GetDeliveryFees(ids []string) (map[string]model.DeliveryFees, error)
	// SaveEvent saves an event to the database.
	SaveEvent(topic string, eventData interface{}, inTx func() error) error
}

//go:generate go-bindata -pkg db -o migrations.go migrations/...
//go:generate mockery --name=DB --output=../mocks
