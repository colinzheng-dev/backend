package server

import (
	"encoding/json"
	"github.com/veganbase/backend/chassis"
	"net/http"
	"time"
)

func (s *Server) sendTestEvent(w http.ResponseWriter, r *http.Request) (interface{}, error) {

	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		return chassis.NotFound(w)
	}

	// Read request body.
	body, err := chassis.ReadBody(r, 0)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	// Unmarshal request data -- validates JSON request body.
	event := chassis.Event{}
	if err = json.Unmarshal(body, &event); err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	event.EventID = chassis.GenerateUUID("evt_test")
	event.Livemode = false
	event.CreatedAt = time.Now()

	if err = s.PubSub.Publish(chassis.WebhookReceiveEventQueue, event); err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	return chassis.NoContent(w)
}
