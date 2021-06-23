package model

import "time"

// Site is the database model for site information.
type Site struct {
	// Unique ID of the site.
	ID string `json:"id" db:"id"`

	// Display name of the site.
	Name string `json:"name" db:"name"`

	// The base URL of the site, e.g. https://ethicalbuzz.com.
	URL string `json:"url" db:"url"`

	// The email domain to use for the site, e.g. ethicalbuzz.com.
	EmailDomain string `json:"email_domain" db:"email_domain"`

	// The signature used in emails from the site.
	Signature string `json:"signature" db:"signature"`

	// The fee charged on each purchase.
	Fee float64 `json:"fee" db:"fee"`

	// Creation timestamp.
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
