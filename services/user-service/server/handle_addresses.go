package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/user-service/db"
	"github.com/veganbase/backend/services/user-service/events"
	"github.com/veganbase/backend/services/user-service/model"
	"googlemaps.github.io/maps"
	_ "googlemaps.github.io/maps"
)

func (s *Server) createAddress(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod != chassis.SessionAuth {
		return chassis.NotFound(w)
	}

	// Read new payout account request body.
	body, err := chassis.ReadBody(r, 0)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	addr := model.Address{}

	if err = json.Unmarshal(body, &addr); err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	//validating fields of payment method
	var zeroTime time.Time
	if addr.CreatedAt != zeroTime {
		return chassis.BadRequest(w, "can't set read-only field 'created_at'")
	}
	if addr.Owner == "" {
		addr.Owner = authInfo.UserID
	} else if addr.Owner != authInfo.UserID {
		return chassis.BadRequest(w, "can't create address for other user")
	}

	if err = s.setGeolocation(&addr); err != nil {
		return nil, err
	}

	if err = s.db.CreateAddress(&addr); err != nil {
		return nil, err
	}
	chassis.Emit(s, events.AddressCreated, addr)

	return addr, err
}

func (s *Server) getDefaultAddress(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod != chassis.SessionAuth {
		return chassis.NotFound(w)
	}
	addr, err := s.db.DefaultAddressByUserId(authInfo.UserID)
	if err != nil {
		if err == db.ErrAddressNotFound {
			return chassis.NotFoundWithMessage(w, "the user does not have a default address")
		}
		return nil, err
	}

	return addr, nil
}

func (s *Server) getAddress(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod != chassis.SessionAuth {
		return chassis.NotFound(w)
	}
	addrId := chi.URLParam(r, "addr_id")
	if addrId == "" {
		return chassis.BadRequest(w, "missing address ID")
	}
	addr, err := s.db.AddressById(addrId)
	if err != nil {
		if err == db.ErrAddressNotFound {
			return chassis.NotFoundWithMessage(w, "address not found")
		}
		return nil, err
	}

	return addr, nil
}

func (s *Server) getAddresses(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod != chassis.SessionAuth {
		return chassis.NotFound(w)
	}

	addrs, err := s.db.AddressesByUserId(authInfo.UserID)
	if err != nil {
		return nil, err
	}

	return addrs, nil
}

func (s *Server) updateAddress(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod != chassis.SessionAuth {
		return chassis.NotFound(w)
	}
	addrId := chi.URLParam(r, "addr_id")
	if addrId == "" {
		return chassis.BadRequest(w, "missing address ID")
	}

	addr, err := s.db.AddressById(addrId)
	if err != nil {
		if err == db.ErrAddressNotFound {
			return chassis.NotFoundWithMessage(w, "address not found")
		}
		return nil, err
	}

	// Read patch request body.
	body, err := chassis.ReadBody(r, 0)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}
	//patching payment method
	if err = addr.Patch(body); err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	if err = s.setGeolocation(addr); err != nil {
		return nil, err
	}
	// Do the update.
	if err = s.db.UpdateAddress(addr); err != nil {
		if err == db.ErrPaymentMethodNotFound {
			return chassis.NotFoundWithMessage(w, "address not found")
		}
		return nil, err
	}

	chassis.Emit(s, events.AddressUpdated, addr)

	return addr, nil
}

func (s *Server) deleteAddress(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Get authentication information from context
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod != chassis.SessionAuth {
		return chassis.NotFound(w)
	}

	addrId := chi.URLParam(r, "addr_id")
	if addrId == "" {
		return chassis.BadRequest(w, "missing address ID")
	}
	// Look up address value.
	addr, err := s.db.AddressById(addrId)
	if err != nil {
		if err == db.ErrAddressNotFound {
			return chassis.NotFoundWithMessage(w, "address not found")
		}
		return nil, err
	}

	// Do the delete.
	if err = s.db.DeleteAddress(addr.ID); err != nil {
		return nil, err
	}
	chassis.Emit(s, events.AddressDeleted, addr)

	return chassis.NoContent(w)
}

//returns an address of a certain user. Ownership is validated.
func (s *Server) getAddressInternal(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	userId := chi.URLParam(r, "user_id")
	if userId == "" {
		return chassis.BadRequest(w, "missing user id")
	}
	addressId := chi.URLParam(r, "addr_id")
	var getDefault bool
	if addressId == "" {
		getDefault = true
	}
	var address *model.Address
	var err error
	if getDefault {
		address, err = s.db.DefaultAddressByUserId(userId)
	} else {
		address, err = s.db.AddressById(addressId)
	}
	if err != nil {
		if err == db.ErrAddressNotFound {
			if getDefault {
				return chassis.NotFoundWithMessage(w, "default address not found")
			}
			return chassis.NotFoundWithMessage(w, "address not found")
		}

		return nil, err
	}
	if address.Owner != userId {
		return chassis.BadRequest(w, "the user is not the owner of the address")
	}

	return address, nil
}

func (s *Server) setGeolocation(addr *model.Address) error {

	r := &maps.GeocodingRequest{
		Address:  buildInlineAddress(addr),
		Language: "en",
		Region:   addr.Country,
	}

	resp, err := s.mapsClient.Geocode(context.Background(), r)
	if err != nil {
		return err
	}

	if len(resp) == 0 {
		return errors.New("geocoder did not return any results")
	}

	addr.Coordinates.Latitude = resp[0].Geometry.Location.Lat
	addr.Coordinates.Longitude = resp[0].Geometry.Location.Lng

	return nil
}

func buildInlineAddress(addr *model.Address) string {
	return addr.HouseNumber + " " + addr.StreetAddress + ",  " + addr.City + ", " + addr.RegionPostalCode
}
