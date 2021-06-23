package server

import (
	"encoding/json"
	"github.com/go-chi/chi"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/user-service/db"
	"github.com/veganbase/backend/services/user-service/events"
	"github.com/veganbase/backend/services/user-service/model"
	"net/http"
	"time"
)
// getCustomer creates a new customer reference of the logged in user. This reference is simply the
// customer_id created by Stripe's API and stored here for future reference (payment service). Only
// one customer is allowed by each user. No changes are allowed, so we must delete it on Stripe and then
// delete the reference here.
func (s *Server) createCustomer(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod != chassis.SessionAuth {
		return chassis.Unauthorized(w, "this resource can only be accessed by authorized users")
	}

	// Read new payout account request body.
	body, err := chassis.ReadBody(r, 0)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	cus := model.Customer{}

	if err = json.Unmarshal(body, &cus); err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	var zeroTime time.Time
	if cus.CreatedAt != zeroTime {
		return chassis.BadRequest(w, "can't set read-only field 'created_at'")
	}
	if cus.UserID == "" {
		cus.UserID = authInfo.UserID
	} else if cus.UserID != authInfo.UserID {
		return chassis.BadRequest(w, "can't create customer for other user")
	}

	//checking if user already has a customer attached
	cusOnDatabase, err := s.db.CustomerByUserId(cus.UserID);
	if err != nil && err != db.ErrCustomerNotFound {
		return nil, err
	}
	if cusOnDatabase != nil {
		return chassis.BadRequest(w, "user already has a customer attached. If you want to change it, delete it on Stripe and here afterwards")
	}

	if err = s.db.CreateCustomer(&cus); err != nil {
		return nil, err
	}
	chassis.Emit(s, events.CustomerCreated, cus)
	return cus, nil
}

//getCustomer gets the customer reference of the logged in user.
func (s *Server) getCustomer(w http.ResponseWriter, r *http.Request) (interface{}, error) {

	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		return chassis.Unauthorized(w, "this resource can only be accessed by authorized users")
	}

	cus, err := s.db.CustomerByUserId(authInfo.UserID)
	if err != nil {
		return nil, err
	}

	return cus, nil
}

//deleteCustomer deletes the customer reference of the logged in user.
func (s *Server) deleteCustomer(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Get authentication information from context
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod != chassis.SessionAuth {
		return chassis.Unauthorized(w, "this resource can only be accessed by authorized users ")
	}

	// Look up payment method value.
	pmt, err := s.db.CustomerByUserId(authInfo.UserID)
	if err != nil {
		if err == db.ErrCustomerNotFound {
			return chassis.NotFoundWithMessage(w, "customer not found")
		}
		return nil, err
	}

	// Do the delete.
	if err = s.db.DeleteCustomer(authInfo.UserID); err != nil {
		return nil, err
	}
	chassis.Emit(s, events.CustomerDeleted, pmt)

	return chassis.NoContent(w)
}


//getCustomerInternal gets the customer reference of a user. This method is unprotected by auth methods, but is
//only available to be called internally via client interface of the service.
func (s *Server) getCustomerInternal(w http.ResponseWriter, r *http.Request) (interface{}, error) {

	userId := chi.URLParam(r, "user_id")
	if userId == "" {
		return chassis.BadRequest(w, "user id missing")
	}

	cus, err := s.db.CustomerByUserId(userId)
	if err != nil {
		return nil, err
	}

	return cus, nil
}
