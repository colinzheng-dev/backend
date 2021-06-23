package server

import (
	"encoding/json"
	"github.com/rs/zerolog/log"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/chassis/pubsub"
	"github.com/veganbase/backend/services/webhook-service/model"
)

const (
	WebhookProcessEventQueue = "wh-process-event-queue"
)

func (s *Server) ReceivedEvents() {
	// Use a single subscription name to process changes by competing consumers.
	ch, _, err := s.PubSub.Subscribe(chassis.WebhookReceiveEventQueue, s.AppName, pubsub.CompetingConsumers)
	if err != nil {
		log.Fatal().Err(err).Msg("couldn't subscribe to receive-webhook topic")
	}

	for {
		data := <-ch
		event := chassis.Event{}

		if err := json.Unmarshal(data, &event); err != nil {
			log.Error().Err(err).Msg("unmarshalling item service event")
		}

		var e model.Event
		e.FromChassisEvent(event)
		if err = s.db.AddEvent(&e); err != nil {
			log.Error().Err(err).Msg("couldn't add event to database")
		}

		if err = s.PubSub.Publish(WebhookProcessEventQueue, event); err != nil {
			log.Error().Err(err).Msg("couldn't add event to pub/sub queue")
		}
	}
}
