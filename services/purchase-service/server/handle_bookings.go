package server

import (
	"github.com/go-chi/chi"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/purchase-service/db"
	"github.com/veganbase/backend/services/purchase-service/events"
	"github.com/veganbase/backend/services/purchase-service/model"
	"net/http"
	"net/url"
)
// get all bookings of a specific purchase
func (s *Server) bookingsByPurchaseSearch(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		return chassis.NotFound(w)
	}

	purId := chi.URLParam(r, "pur_id")
	if purId == "" {
		return chassis.BadRequest(w,"missing purchase ID")
	}

	qs, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return chassis.BadRequest(w, "error while parsing the query")
	}

	params := chassis.Pagination{}
	// Pagination parameters.
	if err := chassis.PaginationParams(qs, &params.Page, &params.PerPage); err != nil {
		return nil, err
	}

	purchase, err := s.db.PurchaseById(purId)
	if err != nil {
		if err == db.ErrPurchaseNotFound {
			return chassis.NotFoundWithMessage(w, "purchase not found")
		}
		return chassis.BadRequest(w, err.Error())
	}

	if authInfo.UserID != purchase.BuyerID {
		return chassis.NotFound(w)
	}

	bookings, total, err := s.db.BookingsByPurchase(purchase.Id, &params)
	if err != nil  {
		return chassis.BadRequest(w, err.Error())
	}
	chassis.BuildPaginationResponse(w, r, params.Page, params.PerPage, *total)
	return s.BuildBookingFullResponse(bookings)
}

// get all bookings of a specific user or host
func (s *Server) bookingsSearch(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		return chassis.NotFound(w)
	}

	qs, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return chassis.BadRequest(w, "error while parsing the query")
	}

	params := chassis.Pagination{}
	var host *string
	// Pagination parameters.
	if err := chassis.PaginationParams(qs, &params.Page, &params.PerPage); err != nil {
		return nil, err
	}

	chassis.StringParam(qs,"host", &host )
	org := qs.Get("org")


	//getting all bookings that the logged in user is the host
	if host != nil && *host == "true" {
		owner := authInfo.UserID
		if org != "" {
			isOrgMember, err := s.userSvc.IsUserOrgMember(authInfo.UserID, org)
			if err != nil {
				return chassis.BadRequest(w, err.Error())
			}
			if !isOrgMember {
				return chassis.BadRequest(w, "user not member of organisation")
			}
			owner = org
		}
		bookings, total, err := s.db.BookingsByHost(owner, &params)
		if err != nil {
			return chassis.BadRequest(w, err.Error())
		}
		chassis.BuildPaginationResponse(w, r, params.Page, params.PerPage, *total)
		return s. BuildBookingFullResponse(bookings)
	}

	//getting all bookings of the current user
	bookings, total, err := s.db.BookingsByBuyer(authInfo.UserID, &params)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}
	chassis.BuildPaginationResponse(w, r, params.Page, params.PerPage, *total)
	return s.BuildBookingFullResponse(bookings)
}

// Handle normal booking search.
func (s *Server) bookingSearch(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		return chassis.NotFound(w)
	}

	bokId := chi.URLParam(r, "bok_id")
	if bokId == "" {
		return chassis.BadRequest(w, "missing booking id")
	}

	//check if booking exists
	booking, err := s.db.BookingById(bokId)
	if err != nil {
		if err == db.ErrOrderNotFound {
			return chassis.NotFoundWithMessage(w, "booking not found")
		}
		return chassis.BadRequest(w, err.Error())
	}

	if authInfo.UserID != booking.BuyerID  {
		return chassis.BadRequest(w, "the user is not the owner of the booking")
	}

	//calling user service to get updated information about users
	userInfo, err := s.userSvc.Info([]string{booking.BuyerID, booking.Host})
	if err != nil {
		return booking, nil
	}

	itemInfo, err := s.itemSvc.GetItemsInfo([]string{booking.ItemID})
	if err != nil {
		return booking, nil
	}

	return *model.GetFullBooking(booking, userInfo[booking.BuyerID], userInfo[booking.Host], itemInfo[booking.ItemID]), err
}

// patchBookingInternal performs a patch operation in a booking item.
// the existence of the booking is evaluated, so as its ownership
func (s *Server) patchBooking(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		return chassis.NotFound(w)
	}

	bokId := chi.URLParam(r, "bok_id")
	if bokId == "" {
		return chassis.BadRequest(w, "missing booking id")
	}

	// Read patch request body.
	body, err := chassis.ReadBody(r, 0)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	// check if the booking exists at the database
	booking, err := s.db.BookingById(bokId)
	if err != nil {
		if err == db.ErrBookingNotFound {
			return chassis.NotFoundWithMessage(w, "booking not found")
		}
		return nil, err
	}

	//check if user is owner of the booking
	if authInfo.UserID != booking.BuyerID {
		return chassis.NotFound(w)
	}

	//patch booking
	if err = booking.Patch(body); err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	// Do the update.
	if err = s.db.UpdateBooking(booking); err != nil {
		return nil, err
	}

	chassis.Emit(s, events.BookingUpdated, booking)

	return booking, nil
}

// patchBookingInternal performs a patch operation in a booking item.
// the existence of the booking is evaluated, so as its ownership
func (s *Server) patchBookingInternal(w http.ResponseWriter, r *http.Request) (interface{}, error) {

	bokId := chi.URLParam(r, "bok_id")
	if bokId == "" {
		return chassis.BadRequest(w, "missing booking id")
	}

	// Read patch request body.
	body, err := chassis.ReadBody(r, 0)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	// check if the booking exists at the database
	booking, err := s.db.BookingById(bokId)
	if err != nil {
		if err == db.ErrBookingNotFound {
			return chassis.NotFoundWithMessage(w, "booking not found")
		}
		return nil, err
	}

	//patch booking
	if err = booking.Patch(body); err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	// Do the update.
	if err = s.db.UpdateBooking(booking); err != nil {
		return nil, err
	}

	chassis.Emit(s, events.BookingUpdated, booking)

	return booking, nil
}


func (s *Server) BuildBookingFullResponse(rawBookings *[]model.Booking) (interface{}, error) {

	if len(*rawBookings) == 0 {
		return rawBookings, nil
	}
	//Getting unique ids to acquire full user information
	userIDs := map[string]bool{}
	itemIDs := map[string]bool{}
	for _, bk := range *rawBookings {
		userIDs[bk.BuyerID] = true
		userIDs[bk.Host] = true
		itemIDs[bk.ItemID] = true
	}

	uniqueUserIDs := []string{}
	for k := range userIDs {
		uniqueUserIDs = append(uniqueUserIDs, k)
	}
	uniqueItemIDs := []string{}
	for k := range itemIDs {
		uniqueItemIDs = append(uniqueItemIDs, k)
	}

	//calling user service to get updated information about users
	userInfo, err := s.userSvc.Info(uniqueUserIDs)
	if err != nil {
		return rawBookings, nil
	}

	itemInfo, err := s.itemSvc.GetItemsInfo(uniqueItemIDs)
	if err != nil {
		return rawBookings, nil
	}

	var fullBookings []model.FullBooking
	for _, bk := range *rawBookings {
		fullBookings = append(fullBookings, *model.GetFullBooking(&bk, userInfo[bk.BuyerID], userInfo[bk.Host], itemInfo[bk.ItemID]))
	}

	return fullBookings, nil

}
