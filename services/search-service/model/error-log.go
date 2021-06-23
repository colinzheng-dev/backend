package model

import "time"

type ErrorLog struct {
	ID        int       `db:"id"`
	Action    string    `db:"action"`
	Error     string    `db:"error"`
	CreatedAt time.Time `db:"created_at"`
}
