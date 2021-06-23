package server

import (
	"github.com/rs/zerolog/log"

	"github.com/veganbase/backend/services/email-service/transform"
)

// Main email sender goroutine: runs off of multiplexed message
// channel.
func (s *Server) sender() {
	// TODO: HANDLE MESSAGES IN PARALLEL WITH RATE LIMITING.
	for {
		ev := <-s.muxCh

		log.Info().
			Str("topic", ev.topic).
			Str("event", string(ev.data)).
			Msg("email send event")
		topicName := ev.topic
		topic, ok := s.topics[topicName]
		if !ok {
			log.Error().
				Str("topic", topicName).
				Msg("unknown email topic")
			continue
		}
		fields := make(map[string]interface{})
		err := transform.Unmarshal(topicName, ev.data, fields)
		if err != nil {
			log.Error().
				Str("topic", topicName).
				Err(err).
				Msg("could't unmarshal message data as map")
			continue
		}
		sitename := fields["site"].(string)
		if sitename == "" {
			sitename = "unknown"
		}
		site := s.siteSvc.Sites()[sitename]
		language := fields["language"].(string)
		if language == "" {
			language = "en"
		}
		//TODO: language is being used for nothing.
		//      check what was the idea behind it
		err = s.mailer.Send(topic, site, language, fields)
		if err != nil {
			log.Error().Err(err).
				Str("topic", topicName).
				Str("site", sitename).
				Str("message", string(ev.data)).
				Msg("mailer couldn't send email")
			continue
		}
		log.Info().
			Str("topic", topicName).
			Str("site", sitename).
			Str("message", string(ev.data)).
			Msg("email sent")
	}
}
