package server

import (
	"encoding/json"
	"github.com/go-chi/chi"
	"github.com/veganbase/backend/services/user-service/db"
	"github.com/veganbase/backend/services/user-service/events"
	"net/url"
	"strings"
	"time"

	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/user-service/model"
	"net/http"
)

func (s *Server) createDeliveryFees(w http.ResponseWriter, r *http.Request) (interface{}, error) {
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

	fees := model.DeliveryFees{}

	if err = json.Unmarshal(body, &fees); err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	if fees.Owner != "" {
		return chassis.BadRequest(w, "can't set read-only field 'owner'")
	}

	var zeroTime time.Time
	if fees.CreatedAt != zeroTime {
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
		fees.Owner = org.ID
	}

	//checking if owner field is an org and if the logged in user is authorized to manipulate it
	if fees.Owner != "" && len(fees.Owner) > 3 {
		if fees.Owner[0:4] == "org_" {
			//checking if logged in user is admin of the org
			if err = s.userIsOrgAdmin(authInfo.UserID, fees.Owner); err != nil {
				return chassis.Forbidden(w)
			}
		}
	}

	if fees.Owner == "" {
		fees.Owner = authInfo.UserID
	}

	//checking if the owner already configured delivery fees
	check, err := s.db.DeliveryFeesByOwner(fees.Owner)
	if check != nil {
		return chassis.BadRequest(w, "delivery fees configuration already exists")
	}
	if err != nil && err != db.ErrDeliveryFeesNotFound{
		return chassis.BadRequest(w, err.Error())
	}

	if err = s.db.CreateDeliveryFees(&fees); err != nil {
		return nil, err
	}

	chassis.Emit(s, events.DeliveryFeesCreated, fees)
	return fees, nil
}

func (s *Server) getUserDeliveryFees(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod != chassis.SessionAuth  {
		return chassis.NotFound(w)
	}

	fees, err := s.db.DeliveryFeesByOwner(authInfo.UserID)
	if err != nil {
		if err == db.ErrDeliveryFeesNotFound {
			return chassis.NotFoundWithMessage(w, err.Error())
		}
		return nil, err
	}

	return fees, nil
}
func (s *Server) getOrgDeliveryFees(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	//check if user can get sensitive information about the org
	org, allowed, err := s.orgModAllowed(w, r, false)
	if !allowed {
		return nil, err
	}

	fees, err := s.db.DeliveryFeesByOwner(org.ID)
	if err != nil {
		if err == db.ErrDeliveryFeesNotFound {
			return chassis.NotFoundWithMessage(w, err.Error())
		}
		return nil, err
	}

	return fees, nil
}

//updateUserDeliveryFees performs user specific validations and updates its delivery fees configuration.
func (s *Server) updateUserDeliveryFees(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod != chassis.SessionAuth  {
		return chassis.NotFound(w)
	}
	return s.updateDeliveryFees(w, r, authInfo.UserID)
}
//updateOrgDeliveryFees performs organisation specific validations and updates its delivery fees configuration.
func (s *Server) updateOrgDeliveryFees(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	org, allowed, err := s.orgModAllowed(w, r, false)
	if !allowed {
		return nil, err
	}

	return s.updateDeliveryFees(w, r, org.ID)
}

//updateDeliveryFees performs database update operation of a specific owner delivery fees configuration
func (s *Server) updateDeliveryFees(w http.ResponseWriter, r *http.Request, owner string) (interface{}, error) {
	fees, err := s.db.DeliveryFeesByOwner(owner)
	if err != nil {
		return nil, err
	}

	// Read patch request body.
	body, err := chassis.ReadBody(r, 0)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}
	//patching account
	if err = fees.Patch(body); err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	// Do the update.
	if err = s.db.UpdateDeliveryFees(fees); err != nil {
		if err == db.ErrDeliveryFeesNotFound {
			return chassis.NotFoundWithMessage(w, err.Error())
		}
		return nil, err
	}

	chassis.Emit(s, events.DeliveryFeesUpdated, fees)

	return fees, nil
}

//deleteUserDeliveryFees performs user specific validations and deletes its delivery fees configuration.
func (s *Server) deleteUserDeliveryFees(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod != chassis.SessionAuth  {
		return chassis.NotFound(w)
	}
	return s.deleteDeliveryFees(w, r, authInfo.UserID)
}

//deleteOrgDeliveryFees performs organisation specific validations and deletes its delivery fees configuration.
func (s *Server) deleteOrgDeliveryFees(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	org, allowed, err := s.orgModAllowed(w, r, false)
	if !allowed {
		return nil, err
	}

	return s.deleteDeliveryFees(w, r, org.ID)
}
//deleteDeliveryFees performs database delete operation of a specific owner delivery fees configuration
func (s *Server) deleteDeliveryFees(w http.ResponseWriter, r *http.Request, owner string) (interface{}, error) {
	// Look up delivery fees configuration.
	fees, err := s.db.DeliveryFeesByOwner(owner)

	if err != nil {
		if err == db.ErrDeliveryFeesNotFound {
			return chassis.NotFoundWithMessage(w, err.Error())
		}
		return nil, err
	}

	// Do the delete.
	if err = s.db.DeleteDeliveryFees(fees.ID); err != nil {
		return nil, err
	}
	chassis.Emit(s, events.DeliveryFeesDeleted, fees)

	return chassis.NoContent(w)
}


func (s *Server) getDeliveryFeesInternal(w http.ResponseWriter, r *http.Request) (interface{}, error) {

	qs, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return chassis.BadRequest(w, "error while parsing the query")
	}
	rawIds := qs.Get("ids")
	if rawIds == ""{
		return chassis.BadRequest(w, "missing ids")
	}

	ids := strings.Split(rawIds, ",")

	fees, err := s.db.GetDeliveryFees(ids)
	if err != nil {
		if err == db.ErrDeliveryFeesNotFound {
			return chassis.NotFoundWithMessage(w, err.Error())
		}
		return nil, err
	}

	return fees, nil
}