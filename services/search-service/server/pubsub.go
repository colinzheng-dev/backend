package server

import (
	"encoding/json"

	"github.com/rs/zerolog/log"
	"github.com/veganbase/backend/chassis/pubsub"
	item_events "github.com/veganbase/backend/services/item-service/events"
)

const subName = "search-service-item-changes"

func (s *Server) handleItemEvents() {
	// Use a single subscription name to process changes by competing
	// consumers.
	ch, _, err := s.PubSub.Subscribe(item_events.ItemChange, subName,
		pubsub.CompetingConsumers)
	if err != nil {
		log.Fatal().Err(err).
			Msg("couldn't subscribe to item change events")
	}

	for {
		data := <-ch
		event := item_events.ItemEvent{}
		err := json.Unmarshal(data, &event)
		if err != nil {
			log.Error().Err(err).
				Msg("unmarshalling item service event")
		}

		switch event.EventType {
		case item_events.ItemCreated, item_events.ItemUpdated:
			s.processItemUpdate(event.ItemID)
		case item_events.ItemDeleted:
			s.processItemDelete(event.ItemID)
		}
	}
}
