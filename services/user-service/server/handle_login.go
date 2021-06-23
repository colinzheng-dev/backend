package server

import (
	"net/http"

	"github.com/pkg/errors"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/user-service/client"
	"github.com/veganbase/backend/services/user-service/events"
	"github.com/veganbase/backend/services/user-service/messages"
)

// Handle login route.
func (s *Server) login(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Unmarshal request body.
	req := messages.LoginRequest{}
	err := chassis.Unmarshal(r.Body, &req)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	// Perform login processing.
	user, new, err := s.db.LoginUser(req.Email, s.avatarGen)
	if err != nil {
		return nil, errors.Wrap(err, "performing login processing")
	}
	if new {
		chassis.Emit(s, events.UserCreated, req)
	}
	chassis.Emit(s, events.UserLogin, map[string]string{"email": req.Email})

	// Return user response for marshalling.
	resp := client.LoginResponse{
		User:    user,
		NewUser: new,
	}
	return &resp, nil
}
