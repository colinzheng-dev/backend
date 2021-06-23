package server

import (
	"encoding/json"
	"github.com/rs/zerolog/log"
	"github.com/veganbase/backend/chassis/pubsub"
	social "github.com/veganbase/backend/services/social-service/events"
	"time"
)

// HandleItemRank subscribes to messages indicating that a new review was made
// so we can call social service to recalculate the ranking.
func (s *Server) HandleItemRank() {
	updCh, _, err := s.PubSub.Subscribe(social.ItemRankTopic, s.AppName, pubsub.Fanout)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to subscribe to user update topic")
	}

	for {
		itemJSON := <-updCh
		var itemId string
		if err := json.Unmarshal(itemJSON, &itemId); err != nil {
			log.Error().Err(err).Msg("decoding itemId message from "+ social.ItemRankTopic)
			continue
		}
		rank, err := s.socialSvc.GetOverallRank(itemId)
		if err != nil {
			log.Error().Err(err).Msg("obtaining item rank from social-service")
			continue
		}
		if err = s.db.SetItemRank(itemId, *rank); err != nil {
			log.Error().Err(err).Msg("updating item rank")
		}
	}
}


func (s *Server) HandleItemUpvotes() {
	//get period from env
	period := time.Tick(10 * time.Minute)
	for range period {
		//query all items that were upvoted
		upvoteInfo, err := s.socialSvc.GetUpvotesCount()
		if err != nil {
			log.Error().Err(err).Msg("obtaining items upvote sum from social-service")
			continue
		}
		for _, ui:= range *upvoteInfo {
			if err = s.db.SetItemUpvotes(ui.ItemId, ui.Quantity); err != nil {
				log.Error().Err(err).Msg("updating item upvote")
			}
		}

	}
}
