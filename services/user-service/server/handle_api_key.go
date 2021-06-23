package server

import (
	"net/http"

	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/user-service/db"
	"github.com/veganbase/backend/services/user-service/events"
	"github.com/veganbase/backend/services/user-service/messages"
)

func (s *Server) createAPIKey(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	userID, _, _ := accessControl(r)
	if userID == nil {
		return chassis.NotFound(w)
	}

	rawKey, rawSecret, encryptedSecret, err := s.generateNewAPIKey()
	if err != nil {
		return nil, err
	}

	if err := s.db.SaveHashedAPIKey(*rawKey, *encryptedSecret, *userID); err != nil {
		if err == db.ErrUserNotFound {
			return chassis.NotFound(w)
		}
		return nil, err
	}

	chassis.Emit(s, events.CreateAPIKey, map[string]string{"user_id": *userID})
	response := messages.APIKeyResponse{
		APIKey: *rawKey,
		APISecret: *rawSecret,
	}
	return &response, nil
}

func (s *Server) generateNewAPIKey() (*string,*string,*string, error) {

	rawKey := chassis.NewBareID(32)
	rawSecret := chassis.NewBareID(32)

	hashedSecret, err :=  chassis.HashAndSalt(rawSecret)
	if err != nil {
		return nil, nil, nil, err
	}

	return &rawKey, &rawSecret, &hashedSecret, nil
}

func (s *Server) deleteAPIKey(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	userID, _, _ := accessControl(r)
	if userID == nil {
		return chassis.NotFound(w)
	}

	err := s.db.DeleteAPIKey(*userID)
	if err == db.ErrUserNotFound {
		return chassis.NotFound(w)
	}
	if err != nil {
		return nil, err
	}

	chassis.Emit(s, events.DeleteAPIKey, map[string]string{"user_id": *userID})
	return chassis.NoContent(w)
}
