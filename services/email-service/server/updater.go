package server

import (
	"time"

	"github.com/rs/zerolog/log"
	"github.com/veganbase/backend/services/email-service/model"
)

// UpdateTopics is a function that runs in a goroutine and regularly
// updates the topics list from the database.
func (s *Server) UpdateTopics() {
	defer log.Error().Msg("dropping out of topic updater")
	s.topicUpdate()
	for range time.Tick(10 * time.Second) {
		s.topicUpdate()
	}
}

// Do topic update from database.
func (s *Server) topicUpdate() {
	dbTopicsTmp, err := s.db.Topics()
	if err != nil {
		log.Error().Err(err).Msg("couldn't read topics from database")
		return
	}
	topicsTmp := map[string]*model.TopicInfo{}
	for _, t := range dbTopicsTmp {
		topicsTmp[t.Name] = t.Info()
	}
	addTopics, removeTopics := topicsDiff(s.topics, topicsTmp)

	for _, t := range removeTopics {
		if err = s.removeTopicSubscription(t); err != nil {
			log.Error().Err(err).
				Str("topic", t).
				Msg("couldn't remove Pub/Sub subscription")
		}
	}
	for _, t := range addTopics {
		if err = s.addTopicSubscription(t); err != nil {
			log.Error().Err(err).
				Str("topic", t).
				Msg("couldn't add Pub/Sub subscription")
		}
	}
	s.setTopics(topicsTmp)
}

// Set current topic list (in its own function for locking purposes).
func (s *Server) setTopics(ts map[string]*model.TopicInfo) {
	s.muTopic.Lock()
	defer s.muTopic.Unlock()
	s.topics = ts
}

// Determine whether two topic lists are different and determine which
// items are added to or removed from the first list to make the
// second.
func topicsDiff(topics1, topics2 map[string]*model.TopicInfo) ([]string, []string) {
	// Make sets of topic names: in first, in second, all in both.
	tall := map[string]*model.TopicInfo{}
	for _, t := range topics1 {
		tall[t.Name] = t
	}
	for _, t := range topics2 {
		tall[t.Name] = t
	}

	// Work out names not in one set but in other (so need to be added
	// or removed).
	add := []string{}
	remove := []string{}
	for t := range tall {
		_, in1 := topics1[t]
		_, in2 := topics2[t]
		if in1 && !in2 {
			remove = append(remove, t)
		}
		if !in1 && in2 {
			add = append(add, t)
		}
	}

	return add, remove
}
