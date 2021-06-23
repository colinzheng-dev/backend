package db

import (
	"errors"
	"time"
)

// LoginTokenDuration is the time for which a login token is valid:
// tokens are considered to have expired if they have not been
// presented within this time of the time they are created.
const LoginTokenDuration = 1 * time.Hour

// ErrLoginTokenNotFound is the error returned when an unknown login
// token is submitted for checking.
var ErrLoginTokenNotFound = errors.New("login token not found")

// ErrSessionNotFound is the error returned when an unknown session ID
// is used.
var ErrSessionNotFound = errors.New("session not found")

// DB describes the database operations used by the API gateway
// (mostly for something that could probably be pulled out as a
// separate authentication service)
type DB interface {
	// CreateLoginToken creates a new unique six-digit numerical login
	// token for the given email. It returns the token directly.
	CreateLoginToken(email string, site string, language string) (string, error)

	// CheckLoginToken checks whether a given login token is valid and
	// has not expired. If the token is good, the email address, site
	// and language associated with it are returned.
	CheckLoginToken(token string) (string, string, string, error)

	// CreateSession generates a new session token and stores it along
	// with the associated user information.
	CreateSession(userID string, userEmail string, userIsAdmin bool) (string, error)

	// UpdateSessions updates all sessions for a user to reflect changes
	// in user data.
	UpdateSessions(userID string, userEmail string, userIsAdmin bool) error

	// LookupSession checks a session token and returns the associated
	// user ID, email and admin flag if the session is known.
	LookupSession(token string) (string, string, bool, error)

	// DeleteSession deletes a single session, i.e. logs a user out of
	// their current session.
	DeleteSession(token string) error

	// DeleteUserSessions deletes all sessions for a user, i.e. logs the
	// user out of all devices where they're logged in.
	DeleteUserSessions(userID string) error

	// SaveEvent saves an event to the database.
	SaveEvent(label string, eventData interface{}, inTx func() error) error
}

//go:generate go-bindata -pkg db -o migrations.go migrations/...
