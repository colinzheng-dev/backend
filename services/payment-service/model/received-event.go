package model

import "time"

type ReceivedEvent struct {
	EventId        string    `db:"event_id"`
	IdempotencyKey string    `db:"idempotency_key"`
	EventType      string    `db:"event_type"`
	IsHandled      bool      `db:"is_handled"`
	CreatedAt      time.Time `db:"created_at"`
}
