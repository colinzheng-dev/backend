package server

import (
	"encoding/json"
	"github.com/go-chi/chi"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/oauth"
	"github.com/veganbase/backend/services/user-service/db"
	"github.com/veganbase/backend/services/user-service/events"
	"time"

	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/user-service/model"
	"net/http"
)

func (s *Server) createPayoutAccount(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod != chassis.SessionAuth  {
		return chassis.NotFound(w)
	}

	idOrSlug := chi.URLParam(r, "id_or_slug")

	// Read new payout account request body.
	body, err := chassis.ReadBody(r, 0)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	acc := model.PayoutAccountRequest{}

	if err = json.Unmarshal(body, &acc); err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	if acc.Code == "" {
		return chassis.BadRequest(w, "code must be provided")
	}

	var zeroTime time.Time
	if acc.CreatedAt != zeroTime {
		return chassis.BadRequest(w, "can't set read-only field 'created_at'")
	}

	if idOrSlug != "" {
		org, err := s.db.OrgByIDorSlug(idOrSlug)
		if err != nil {
			if err == db.ErrOrgNotFound {
				return chassis.NotFoundWithMessage(w, err.Error())
			}
			return nil, err
		}
		acc.Owner = org.ID
	}

	//checking if owner field is an org and if the logged in user is authorized to manipulate it
	if acc.Owner != "" && len(acc.Owner) > 3 {
		if acc.Owner[0:4] == "org_" {
			//checking if logged in user is admin of the org
			if err = s.userIsOrgAdmin(authInfo.UserID, acc.Owner); err != nil {
				return chassis.Forbidden(w)
			}
		}
	}

	if acc.Owner == "" {
		acc.Owner = authInfo.UserID
	}

	//checking if the owner already has a payout account
	check, err := s.db.PayoutAccountByOwner(acc.Owner)
	if check != nil {
		return chassis.BadRequest(w, "payout account already exists")
	}
	if err != nil && err != db.ErrPayoutAccountNotFound{
		return chassis.BadRequest(w, err.Error())
	}

	//getting account information
	auth, err := s.fetchUserCredentials(acc.Code)
	if err != nil {
		return chassis.BadRequest(w, "error while acquiring user-credentials on Stripe")
	}

	//setting account number and saving
	acc.AccountNumber = auth.StripeUserID
	if err = s.db.CreatePayoutAccount(&acc.PayoutAccount); err != nil {
		return nil, err
	}

	chassis.Emit(s, events.PayoutAccountCreated, acc.PayoutAccount)
	return acc.PayoutAccount, nil
}

func (s *Server) getUserPayoutAccount(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod != chassis.SessionAuth  {
		return chassis.NotFound(w)
	}

	acc, err := s.db.PayoutAccountByOwner(authInfo.UserID)
	if err != nil {
		if err == db.ErrPayoutAccountNotFound {
			return chassis.NotFoundWithMessage(w, err.Error())
		}
		return nil, err
	}

	return acc, nil
}
func (s *Server) getOrgPayoutAccount(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Check whether the modification is allowed: user is administrator
	// or is organisation administrator for the organisation.
	org, allowed, err := s.orgModAllowed(w, r, false)
	if !allowed {
		return nil, err
	}

	acc, err := s.db.PayoutAccountByOwner(org.ID)
	if err != nil {
		if err == db.ErrPayoutAccountNotFound {
			return chassis.NotFoundWithMessage(w, err.Error())
		}
		return nil, err
	}

	return acc, nil
}


//updateUserPayoutAccount performs user specific validations and updates the payout account.
func (s *Server) updateUserPayoutAccount(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod != chassis.SessionAuth  {
		return chassis.NotFound(w)
	}
	return s.updatePayoutAccount(w, r, authInfo.UserID)
}
//updateOrgPayoutAccount performs organisation specific validations and updates the payout account.
func (s *Server) updateOrgPayoutAccount(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	org, allowed, err := s.orgModAllowed(w, r, false)
	if !allowed {
		return nil, err
	}

	return s.updatePayoutAccount(w, r, org.ID)
}

func (s *Server) updatePayoutAccount(w http.ResponseWriter, r *http.Request, owner string) (interface{}, error) {

	acc, err := s.db.PayoutAccountByOwner(owner)
	if err != nil {
		return nil, err
	}

	// Read patch request body.
	body, err := chassis.ReadBody(r, 0)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}
	//patching account
	if err = acc.Patch(body); err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	// Do the update.
	if err = s.db.UpdatePayoutAccount(acc); err != nil {
		if err == db.ErrPayoutAccountNotFound {
			return chassis.NotFoundWithMessage(w, err.Error())
		}
		return nil, err
	}

	chassis.Emit(s, events.PayoutAccountUpdated, acc)

	return acc, nil
}

//deleteUserPayoutAccount performs user specific validations and deletes the payout account.
func (s *Server) deleteUserPayoutAccount(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod != chassis.SessionAuth  {
		return chassis.NotFound(w)
	}
	return s.deletePayoutAccount(w, r, authInfo.UserID)
}
//deleteOrgPayoutAccount performs organisation specific validations and deletes the payout account.
func (s *Server) deleteOrgPayoutAccount(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	org, allowed, err := s.orgModAllowed(w, r, false)
	if !allowed {
		return nil, err
	}

	return s.deletePayoutAccount(w, r, org.ID)
}

func (s *Server) deletePayoutAccount(w http.ResponseWriter, r *http.Request, owner string) (interface{}, error) {
	// Look up payout account value.
	acc, err := s.db.PayoutAccountByOwner(owner)

	if err != nil {
		if err == db.ErrPayoutAccountNotFound {
			return chassis.NotFoundWithMessage(w, err.Error())
		}
		return nil, err
	}

	// Do the delete.
	if err = s.db.DeletePayoutAccount(acc.ID); err != nil {
		return nil, err
	}
	chassis.Emit(s, events.PayoutAccountDeleted, acc)

	return chassis.NoContent(w)
}

//getPayoutInternal retrieves the payout account of a owner. This method is used internally
// (by payment-service) and there is no authentication nor ownership validations.
func (s *Server) getPayoutInternal(w http.ResponseWriter, r *http.Request) (interface{}, error) {

	ownerId := chi.URLParam(r, "id")
	if ownerId == "" {
		return chassis.BadRequest(w, "missing owner ID")
	}

	acc, err := s.db.PayoutAccountByOwner(ownerId)
	if err != nil {
		if err == db.ErrPayoutAccountNotFound {
			return chassis.NotFoundWithMessage(w, err.Error())
		}
		return nil, err
	}

	return acc, nil
}

//fetchUserCredentials calls stripe to obtain the connected account credentials.
func (s *Server) fetchUserCredentials(code string) (*stripe.OAuthToken, error) {
	stripe.Key = s.stripeKey

	params := &stripe.OAuthTokenParams{
		GrantType: stripe.String("authorization_code"),
		Code:      stripe.String(code),
	}
	return oauth.New(params)

}