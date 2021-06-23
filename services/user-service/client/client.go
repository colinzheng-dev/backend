package client

import (
	"github.com/veganbase/backend/services/user-service/model"
)

// LoginResponse is a structure representing the response body for
// login requests: this is the user profile information plus a flag to
// mark whether this is a first login for a new user.
type LoginResponse struct {
	*model.User
	NewUser bool `json:"new_user,omitempty"`
}

//go:generate mockery --name=Client --output=../mocks
// Client is the service client API for the user service.
type Client interface {
	Login(email, site, language string) (*LoginResponse, error)
	Info(ids []string) (map[string]*model.Info, error)
	OrgsForUser(id string) (map[string]bool, error)
	IsUserOrgMember(userID string, orgID string) (bool, error)
	IsUserOrgAdmin(userID string, orgID string) (bool, error)
	GetCustomer(userID string) (*model.Customer, error)
	GetPayoutAccount(ownerId string) (*model.PayoutAccount, error)
	GetDefaultPaymentMethod(userID string) (*model.PaymentMethod, error)
	GetAddress(userId, addressId string) (*model.Address, error)
	GetDefaultAddress(userId string) (*model.Address, error)
	GetNotificationInfo(userId string) (*model.EmailNotificationInfo, error)
	GetUserByApiKey(apiKey, apiSecret string) (*model.User, error)
	GetDeliveryFees(ids []string) (*map[string]model.DeliveryFees, error)
	GetSSOSecret(orgIDorSlug string) (*string, error)
}
