package model

import "time"

// Topic is a registered email topic, used to determine which Pub/Sub
// topics to listen to for email events.
type Topic struct {
	// Topic ID.
	ID int `db:"id"`

	// Topic name.
	Name string `db:"name"`

	// Send address (e.g. "login" means use "login@site-domain.com").
	SendAddress string `db:"send_address"`

	// Creation timestamp.
	CreatedAt time.Time `db:"created_at"`
}
