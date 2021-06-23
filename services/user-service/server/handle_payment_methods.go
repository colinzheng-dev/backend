package server

import (
	"encoding/json"
	"github.com/go-chi/chi"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/customer"
	"github.com/stripe/stripe-go/paymentmethod"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/user-service/db"
	"github.com/veganbase/backend/services/user-service/events"
	"github.com/veganbase/backend/services/user-service/model"
	"net/http"
	"time"
)

func (s *Server) createPaymentMethod(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod != chassis.SessionAuth {
		return chassis.Unauthorized(w, "this resource can only be accessed by authorized users")
	}

	// Read new payout account request body.
	body, err := chassis.ReadBody(r, 0)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	pmt := model.PaymentMethod{}

	if err = json.Unmarshal(body, &pmt); err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	//validating fields of payment method
	var zeroTime time.Time
	if pmt.CreatedAt != zeroTime {
		return chassis.BadRequest(w, "can't set read-only field 'created_at'")
	}
	if pmt.UserID == "" {
		pmt.UserID = authInfo.UserID
	} else if pmt.UserID != authInfo.UserID {
		return chassis.BadRequest(w, "can't create payment method for other user")
	}
	var cus *model.Customer
	//checking customer and if empty, creating a new one
	cus, err = s.db.CustomerByUserId(authInfo.UserID)
	if err == db.ErrCustomerNotFound {
		usr, err := s.db.UserByID(authInfo.UserID)
		if err != nil {
			return chassis.BadRequest(w, "can't get user")
		}

		stripeCustomer, err := s.createCustomerOnStripe(usr.Email)
		if err != nil {
			return chassis.BadRequest(w, "can't create customer on stripe: "+err.Error())
		}

		cus = &model.Customer{
			CustomerID: stripeCustomer.ID,
			UserID: usr.ID,
		}

		if err = s.db.CreateCustomer(cus); err != nil {
			return chassis.BadRequest(w, "can't create customer our database: "+err.Error())
		}
		chassis.Emit(s, events.CustomerCreated, cus.CustomerID)
	} else if err != nil {
		return nil, err
	}

	//creating payment method on stripe
	_, err = s.attachCustomerToPaymentMethodOnStripe(pmt.PaymentMethodID, cus.CustomerID)
	if err != nil {
		return chassis.BadRequest(w, "can't attach payment method to customer on Stripe: "+err.Error())
	}

	//if new payment method is set to be the default method, flip the flag of the current one
	if pmt.IsDefault == true {
		//checks if user already has a default payment method
		check, err := s.db.DefaultPaymentMethodByUserId(pmt.UserID)
		if err != nil && err != db.ErrPaymentMethodNotFound {
			return chassis.BadRequest(w, "can't validate if user already has a default payment method")
		}
		// if positive, flip isDefault flag and update
		if check != nil {
			check.IsDefault = false
			if err = s.db.UpdatePaymentMethod(check); err != nil {
				return nil, err
			}
		}
	}

	if err = s.db.CreatePaymentMethod(&pmt); err != nil {
		return nil, err
	}
	chassis.Emit(s, events.PaymentMethodCreated, pmt)
	return pmt, err
}

func (s *Server) getDefaultPaymentMethod(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod != chassis.SessionAuth {
		return chassis.Unauthorized(w, "this resource can only be accessed by authorized users")
	}
	pmt, err := s.db.DefaultPaymentMethodByUserId(authInfo.UserID)
	if err != nil {
		if err == db.ErrPaymentMethodNotFound {
			return chassis.NotFoundWithMessage(w, "the user does not have a default payment method")
		}
		return nil, err
	}

	return pmt, nil
}

func (s *Server) getPaymentMethod(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod != chassis.SessionAuth {
		return chassis.Unauthorized(w, "this resource can only be accessed by authorized users")
	}
	pmtId := chi.URLParam(r, "pmt_id")
	if pmtId == "" {
		return chassis.BadRequest(w, "missing payment method ID")
	}
	pmt, err := s.db.PaymentMethodById(pmtId)
	if err != nil {
		return nil, err
	}

	return pmt, nil
}

func (s *Server) getPaymentMethods(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod != chassis.SessionAuth {
		return chassis.Unauthorized(w, "this resource can only be accessed by authorized users")
	}

	pmt, err := s.db.PaymentMethodsByUserId(authInfo.UserID)
	if err != nil {
		return nil, err
	}

	return pmt, nil
}

func (s *Server) updatePaymentMethod(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod != chassis.SessionAuth {
		return chassis.Unauthorized(w, "this resource can only be accessed by authorized users")
	}
	pmtId := chi.URLParam(r, "pmt_id")
	if pmtId == "" {
		return chassis.BadRequest(w, "missing payment method ID")
	}

	pmt, err := s.db.PaymentMethodById(pmtId)
	if err != nil {
		if err == db.ErrPaymentMethodNotFound {
			return chassis.NotFoundWithMessage(w, "payment method not found")
		}
		return nil, err
	}

	// Read patch request body.
	body, err := chassis.ReadBody(r, 0)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}
	//patching payment method
	if err = pmt.Patch(body); err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	//checks if user already has a default payment method
	if pmt.IsDefault == true {
		check, err := s.db.DefaultPaymentMethodByUserId(pmt.UserID)
		if err != nil && err != db.ErrPaymentMethodNotFound {
			return chassis.BadRequest(w, "can't validate if user already has a default payment method")
		}
		// if positive, flip isDefault flag and update
		if check != nil {
			check.IsDefault = false
			if err = s.db.UpdatePaymentMethod(check); err != nil {
				return nil, err
			}
		}
	}

	// Do the update.
	if err = s.db.UpdatePaymentMethod(pmt); err != nil {
		if err == db.ErrPaymentMethodNotFound {
			return chassis.NotFoundWithMessage(w, "payment method not found")
		}
		return nil, err
	}

	chassis.Emit(s, events.PaymentMethodUpdated, pmt)

	return pmt, nil
}

func (s *Server) deletePaymentMethod(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Get authentication information from context
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod != chassis.SessionAuth {
		return chassis.Unauthorized(w, "this resource can only be accessed by authorized users ")
	}

	pmtId := chi.URLParam(r, "pmt_id")
	if pmtId == "" {
		return chassis.BadRequest(w, "missing payment method ID")
	}
	// Look up payment method value.
	pmt, err := s.db.PaymentMethodById(pmtId)
	if err != nil {
		if err == db.ErrPaymentMethodNotFound {
			return chassis.NotFoundWithMessage(w, "payment method not found")
		}
		return nil, err
	}

	// Do the delete.
	if err = s.db.DeletePaymentMethod(pmt.ID); err != nil {
		return nil, err
	}
	chassis.Emit(s, events.PaymentMethodDeleted, pmt)

	return chassis.NoContent(w)
}

func (s *Server) getDefaultPaymentMethodInternal(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	userId := chi.URLParam(r, "user_id")
	if userId == "" {
		return chassis.BadRequest(w, "missing user id")
	}

	pmt, err := s.db.DefaultPaymentMethodByUserId(userId)
	if err != nil {
		if err == db.ErrPaymentMethodNotFound {
			return chassis.NotFoundWithMessage(w, "the user does not have a default payment method")
		}
		return nil, err
	}

	return pmt, nil
}

func (s *Server) createCustomerOnStripe(email string) (*stripe.Customer, error) {
	stripe.Key = s.stripeKey

	params := &stripe.CustomerParams{
		Email: stripe.String(email),
	}
	return customer.New(params)

}

func (s *Server) attachCustomerToPaymentMethodOnStripe(pmId, customerId string) (*stripe.PaymentMethod, error) {
	stripe.Key = s.stripeKey

	//attaching customer to payment method
	attachParams := &stripe.PaymentMethodAttachParams{
		Customer: stripe.String(customerId),
	}
	return paymentmethod.Attach(pmId, attachParams)
}