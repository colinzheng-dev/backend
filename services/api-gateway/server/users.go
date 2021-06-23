package server

import (
	"encoding/json"

	"github.com/rs/zerolog/log"

	"github.com/veganbase/backend/chassis/pubsub"
	user_events "github.com/veganbase/backend/services/user-service/events"
	user_model "github.com/veganbase/backend/services/user-service/model"
)

// HandleUserUpdates subscribes to messages indicating changes in user
// details and passes them on to the sessions table as necessary.
func (s *Server) HandleUserUpdates() {
	updCh, _, err := s.PubSub.Subscribe(user_events.UserUpdated, s.AppName, pubsub.Fanout)
	if err != nil {
		log.Fatal().Err(err).
			Msg("unable to subscribe to user update topic")
	}

	for {
		userJSON := <-updCh
		user := user_model.User{}
		if err := json.Unmarshal(userJSON, &user); err != nil {
			log.Error().Err(err).
				Msg("decoding user update message")
			continue
		}
		if err := s.db.UpdateSessions(user.ID, user.Email, user.IsAdmin); err != nil {
			log.Error().Err(err).
				Msg("updating sessions on user update")
		}
	}
}

// HandleUserDeletions subscribes to messages indicating deleted users
// removes sessions from the sessions table as necessary.
func (s *Server) HandleUserDeletions() {
	delCh, _, err := s.PubSub.Subscribe(user_events.UserDeleted, s.AppName, pubsub.Fanout)
	if err != nil {
		log.Fatal().Err(err).
			Msg("unable to subscribe to user deleted topic")
	}

	for {
		delJSON := <-delCh
		del := struct {
			UserID string `json:"user_id"`
		}{}
		if err := json.Unmarshal(delJSON, &del); err != nil {
			log.Error().Err(err).
				Msg("decoding user delete message")
			continue
		}
		if err := s.db.DeleteUserSessions(del.UserID); err != nil {
			log.Error().Err(err).
				Msg("deleting sessions on user delete")
		}
	}
}
