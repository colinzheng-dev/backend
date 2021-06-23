package server

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/rs/zerolog/log"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/chassis/pubsub"
	"github.com/veganbase/backend/services/webhook-service/db"
	"github.com/veganbase/backend/services/webhook-service/model"
	"net/http"
	"time"
)



// HandleWebhookMessages subscribes to messages indicating new events that should be sent to webhooks
func (s *Server) HandleWebhookMessages() {
	whEvent, _, err := s.PubSub.Subscribe(WebhookProcessEventQueue, s.AppName, pubsub.CompetingConsumers)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to subscribe to webhooks topic to process events")
	}

	for {
		rawEvent := <-whEvent
		var ev chassis.Event
		if err := json.Unmarshal(rawEvent, &ev); err != nil {
			log.Error().Err(err).Msg("decoding event message")
			continue
		}
		isHandled, err := s.db.IsEventHandled(ev.Destination, ev.Type)
		if err != nil {
			log.Error().Err(err).Msg("checking if event is handled")
			continue
		}
		//checking if event should be handled for this destination
		if !*isHandled {
			log.Info().Msgf("event type %s is not handled", ev.Type)
			if err = s.db.DisableRetryFlag(ev.EventID); err != nil {
				if err == db.ErrEventNotFound {
					log.Error().Err(err).Msgf("couldn't disable retry flag on event %s", ev.EventID)
					continue
				}
				log.Error().Err(err).Msgf("couldn't disable retry flag on event %s", ev.EventID)
			}
		}

			////TODO: USER MAY WANT TO RECEIVE THE EVENT IN MORE THAN ONE URL
			webHook, err := s.db.WebhookByOwnerAndEventType(ev.Destination, ev.Type)
			if err != nil {
				if err == db.ErrWebhookNotFound {
					//webhook not found and message shouldn't be sent
					if err = s.db.DisableRetryFlag(ev.EventID); err != nil {
						if err == db.ErrEventNotFound {
							log.Error().Err(err).Msgf("couldn't disable retry flag on event %s", ev.EventID)
							continue
						}
						log.Error().Err(err).Msgf("couldn't disable retry flag on event %s", ev.EventID)
					}
					continue
				}
				log.Error().Err(err).Msg("getting user webhook settings")
			}

			if err := s.SendEvent(ev, webHook); err != nil {
				log.Error().Err(err).Msg("sending event")
			}

	}
}

func (s *Server) retrySendEvents() {
	for {
		events, err := s.db.PendingEvents()
		if err != nil {
			log.Error().Err(err)
			continue
		}
		for _, e := range *events {
			webHook, err := s.db.WebhookByOwnerAndEventType(e.Destination, e.Type)
			if err != nil {
				if err == db.ErrWebhookNotFound {
					//webhook not found and message shouldn't be sent
					if err = s.db.DisableRetryFlag(e.EventID); err != nil {
						if err == db.ErrEventNotFound {
							log.Error().Err(err).Msgf("couldn't disable retry flag on event %s", e.EventID)
							continue
						}
						log.Error().Err(err).Msgf("couldn't disable retry flag on event %s", e.EventID)
					}
					continue
				}
				log.Error().Err(err).Msg("getting user webhook settings")
			}

			if err := s.SendEvent(e.ToChassisEvent(), webHook); err != nil {
				log.Error().Err(err).Msg("sending event")
			}
		}
		chassis.Wait(2, time.Minute)
	}
}

func (s *Server) SendEvent(event chassis.Event, webhook *model.Webhook) error {
	event.Destination = "" //clearing destination because it is not needed outside veganbase
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	r, err := http.NewRequest(http.MethodPost, webhook.URL, bytes.NewBuffer(payload))
	if err != nil {
		log.Error().Err(err).Msgf("building http request for event %s", event.EventID)
		if err = s.db.IncreaseFailedAttempts(event.EventID); err != nil {
			log.Error().Err(err).Msgf("couldn't increment attempts of sending event %s", event.EventID)
			return err
		}
		return err
	}

	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("Accept", "application/json")
	r.Header.Add("X-Veganbase-Signature", s.BuildHMACSignature(event.Data, webhook.Secret))
	r.Header.Add("X-Veganbase-Event-Type", event.Type)

	client := &http.Client{}
	rsp, err := client.Do(r)
	if err != nil {
		return err
	}

	//All 2xx family of status codes are considered as ACK
	if rsp.StatusCode >= http.StatusOK && rsp.StatusCode <= http.StatusIMUsed {
		if err = s.db.SetSentStatus(event.EventID); err != nil {
			log.Error().Err(err).Msgf("couldn't set sent status on event %s", event.EventID)
			return err
		}
		return nil
	}
	//if error, increase the number of attempts
	if err = s.db.IncreaseFailedAttempts(event.EventID); err != nil {
		log.Error().Err(err).Msgf("couldn't increment attempts of sending event %s", event.EventID)
		return err
	}
	return errors.New("event " + event.EventID + " couldn't be sent, a routine will retry later")
}

func (s *Server) BuildHMACSignature(message json.RawMessage, secret string) string {

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}