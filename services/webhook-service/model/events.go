package model

import (
	"encoding/json"
	"github.com/veganbase/backend/chassis"
	"time"
)

type Event struct {
	EventID      string          `json:"event_id" db:"event_id"`
	Destination  string          `json:"destination" db:"destination"`
	Type         string          `json:"type" db:"type"`
	Livemode     bool            `json:"livemode" db:"livemode"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
	Payload      json.RawMessage `json:"payload" db:"payload"`
	Sent         bool            `json:"sent" db:"sent"`
	SentAt       *time.Time      `json:"sent_at" db:"sent_at"`
	Attempts     int             `json:"attempts" db:"attempts"`
	Retry        bool            `json:"retry" db:"retry"`
	LastRetry    *time.Time      `json:"last_retry" db:"last_retry"`
	BackoffUntil *time.Time      `json:"backoff_until" db:"backoff_until"`
	ReceivedAt   time.Time       `json:"received_at" db:"received_at"`
}

func (e *Event) FromChassisEvent(event chassis.Event) {
	e.EventID = event.EventID
	e.Destination = event.Destination
	e.Type = event.Type
	e.Livemode = event.Livemode
	e.CreatedAt = event.CreatedAt
	e.Payload = event.Data
}

func (e *Event) ToChassisEvent() chassis.Event {
	return chassis.Event{
		EventID:     e.EventID,
		Destination: e.Destination,
		Type:        e.Type,
		Livemode:    e.Livemode,
		CreatedAt:   e.CreatedAt,
		Data:        e.Payload,
	}
}