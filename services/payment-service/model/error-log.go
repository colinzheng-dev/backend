package model

import "time"

type ErrorLog struct {
	EventId   string    `db:"event_id"`
	Error     string    `db:"error"`
	CreatedAt time.Time `db:"created_at"`
}
