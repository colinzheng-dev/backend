package server

import (
	"github.com/rs/zerolog/log"

	"github.com/veganbase/backend/chassis/pubsub"
)

// Remove a subscription for a topic, terminating the goroutine
// processing the subscription's messages and cancelling the
// subscription.
func (s *Server) removeTopicSubscription(topic string) error {
	if closer, ok := s.subClosers[topic]; ok {
		log.Info().
			Str("topic", topic).
			Msg("removing topic subscription")
		close(closer)
		return nil
	}
	return pubsub.ErrUnknownTopic
}

// Add a subscription for a topic. Messages arriving on the topic are
// multiplexed onto a single channel for processing by a goroutine.
func (s *Server) addTopicSubscription(topic string) error {
	// Create subscription.
	log.Info().
		Str("topic", topic).
		Msg("adding topic subscription")
	ch, cancel, err := s.PubSub.Subscribe(topic, s.AppName,
		pubsub.CompetingConsumers)
	if err != nil {
		log.Error().Err(err).
			Str("topic", topic).
			Msg("couldn't subscribe to topic")
	}
	closer := make(chan bool)
	go s.processSubscription(topic, ch, closer, cancel)
	s.muSubs.Lock()
	defer s.muSubs.Unlock()
	s.subClosers[topic] = closer
	return nil
}

// Runs in a goroutine to collect values from subscription channel and
// multiplex them onto server's main incoming processing queue.
func (s *Server) processSubscription(topic string,
	ch chan []byte, done chan bool, cancel func()) {
	for {
		select {
		case msg := <-ch:
			s.muxCh <- subEvent{topic, msg}
		case <-done:
			cancel()
			s.muSubs.Lock()
			defer s.muSubs.Unlock()
			delete(s.subClosers, topic)
			break
		}
	}
}
