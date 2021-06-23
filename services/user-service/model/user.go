package model

import (
	"time"
)

// User is the database model for all user accounts.
// TODO: ADD FIELD VALIDATION.
// TODO: ADD PAYMENT METHODS.
type User struct {
	// Unique ID of the user.
	ID string `json:"id" db:"id"`

	// The user's email address that they used to log into their account.
	Email string `json:"email" db:"email"`

	// The full name of the user.
	Name *string `json:"name" db:"name"`

	// The display name of the user, i.e. the public name that appears
	// in the front end when the user is referenced.
	DisplayName *string `json:"display_name" db:"display_name"`

	// A link to an avatar image for the user.
	// TODO: MAKE THE FOLLOWING AN INSTANCE OF A WebImage TYPE FROM
	// THE CHASSIS, WHICH IS A SYNONYM FOR string WITH AN APPROPRIATE
	// Validate METHOD.
	Avatar *string `json:"avatar,omitempty" db:"avatar"`

	// The user's country as a two-letter ISO country code.
	// TODO: IMPLEMENT A TYPE FOR COUNTRY CODES?
	Country *string `json:"country,omitempty" db:"country"`

	// Is the user an administrator?
	IsAdmin bool `json:"is_admin" db:"is_admin"`

	// Time of last login (never null because user accounts are created
	// only when the user first logs in).
	LastLogin time.Time `json:"last_login" db:"last_login"`

	// The user's API key (which may be empty).
	APIKey *string `json:"api_key,omitempty" db:"api_key"`

	// The user's API secret key (which may be empty).
	APISecretKey *string `json:"secret_key,omitempty" db:"secret_key"`

}
