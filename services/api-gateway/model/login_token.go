package model

import "time"

// LoginToken holds information associating a unique six-digit login
// token with a user email. Login tokens have an expiry time beyond
// which they are no longer considered valid.
type LoginToken struct {
	Token     string    `db:"token"`
	Email     string    `db:"email"`
	Site      string    `db:"site"`
	Language  string    `db:"language"`
	ExpiresAt time.Time `db:"expires_at"`
}
