package chassis

import (
	"encoding/json"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/types"
	"github.com/rs/zerolog/log"
)

const (
	WebhookReceiveEventQueue = "wh-receive-event-queue"
)
// EventStream is an interface representing a sink for application events.
type EventStream interface {
	Publish(topic string, eventData interface{}) error
	SaveEvent(topic string, eventData interface{}, inTx func() error) error
}

// SaveEvent writes an event to a service's database using a standardised format.
func SaveEvent(tx *sqlx.Tx, topic string, eventData interface{}, inTx func() error) error {
	d, _ := json.Marshal(eventData)
	_, err := tx.Exec(`INSERT INTO events (label, event_data) VALUES ($1, $2)`,
		topic, types.JSONText(d))
	if err == nil && inTx != nil {
		err = inTx()
	}
	return err
}

// Emit emits an event, writing it to the local event log (stored in
// the "events" table in the service database) and sending it out on
// the Pub/Sub service.
func Emit(str EventStream, topic string, eventData interface{}) error {
	// Save event to database and publish within a single transaction.
	err := str.SaveEvent(topic, eventData, func() error {
		// Publish to Pub/Sub within transaction.
		return str.Publish(topic, eventData)
	})
	if err != nil {
		d, _ := json.Marshal(eventData)
		log.Error().
			Str("event-topic", topic).
			Str("event-data", string(d)).
			Msg("failed writing event to database")
	}
	return err
}

type Event struct {
	EventID     string          `json:"event_id"`
	Destination string          `json:"destination,omitempty"`
	Type        string          `json:"type"`
	Livemode    bool            `json:"livemode"`
	CreatedAt   time.Time       `json:"created_at"`
	Data        json.RawMessage `json:"data"`
}

//TriggerWebhookEvent generates an Event and trigger it to the topic and saves on database in one transaction.
func TriggerWebhookEvent(str EventStream, destination, eventType string, live bool, message json.RawMessage) error {
	event := Event{
		EventID:     GenerateUUID("evt"),
		Destination: destination,
		Type:        eventType,
		Livemode:    live,
		CreatedAt:   time.Now(),
		Data:        message,
	}

	return Emit(str, WebhookReceiveEventQueue, event)
}

