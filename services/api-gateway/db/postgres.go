package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/api-gateway/model"

	// Import Postgres DB driver.
	_ "github.com/lib/pq"
)

// PGClient is a wrapper for the user database connection.
type PGClient struct {
	DB *sqlx.DB
}

// NewPGClient creates a new user database connection.
func NewPGClient(ctx context.Context, dbURL string) (*PGClient, error) {
	db, err := chassis.DBConnect(ctx, "user", dbURL, Asset, AssetDir)
	if err != nil {
		return nil, err
	}
	return &PGClient{db}, nil
}

// CreateLoginToken creates a new unique six-digit numerical login
// token for the given email. It returns the token directly.
func (pg *PGClient) CreateLoginToken(email string, site string, language string) (string, error) {
	// We deliberately don't validate the email address here. Email
	// addresses are a mess, and the simplest thing to do is just to
	// send an email to the address. If it doesn't work, then the entry
	// in the login_tokens table we create here will get cleaned up
	// after the token expires, and no harm done.

	// Do this in a transaction, since we want to clear out existing
	// tokens for the email address and ensure that we respect the token
	// uniqueness constraint.
	tx, err := pg.DB.Beginx()
	if err != nil {
		return "", err
	}
	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			tx.Rollback()
		}
	}()

	// Delete any existing tokens for this email, and at the same time
	// delete any expired tokens.
	_, err = tx.Exec(deleteTokenForEmail, email)
	if err != nil {
		return "", err
	}

	// Create and insert a unique token.
	token := RandToken()
	for {
		result, err := tx.Exec(insertToken,
			token, email, site, language, time.Now().Add(LoginTokenDuration))
		if err != nil {
			return "", err
		}
		rows, err := result.RowsAffected()
		if err != nil {
			return "", err
		}
		if rows == 1 {
			break
		}

		// Token collision: try again...
		token = RandToken()
	}

	return token, nil
}

const deleteTokenForEmail = `
DELETE FROM login_tokens
 WHERE email = $1 OR expires_at < NOW()`

const insertToken = `
INSERT INTO login_tokens (token, email, site, language, expires_at)
    VALUES ($1, $2, $3, $4, $5)
    ON CONFLICT DO NOTHING`

// CheckLoginToken checks whether a given login token is valid and has
// not expired. If the token is good, the email address associated
// with it is returned.
func (pg *PGClient) CheckLoginToken(token string) (string, string, string, error) {
	// In a transaction...
	tx, err := pg.DB.Beginx()
	if err != nil {
		return "", "", "", err
	}
	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			tx.Rollback()
		}
	}()

	// Look up token.
	tokInfo := TokenInfo{}
	err = tx.Get(&tokInfo, lookupToken, token)
	if err == sql.ErrNoRows {
		return "", "", "", ErrLoginTokenNotFound
	}
	if err != nil {
		return "", "", "", err
	}

	// Clear token entries.
	tx.Exec(cleanupTokens, token)

	return tokInfo.Email, tokInfo.Site, tokInfo.Language, nil
}

// TokenInfo is a temporary structure for holding information about
// the login token being retrieved.
type TokenInfo struct {
	Email    string `db:"email"`
	Site     string `db:"site"`
	Language string `db:"language"`
}

const lookupToken = `
SELECT email, site, language FROM login_tokens
 WHERE token = $1 AND expires_at >= NOW()`

const cleanupTokens = `
DELETE FROM login_tokens WHERE token = $1 OR expires_at < NOW()`

// CreateSession generates a new session token and stores it along
// with the associated user information.
func (pg *PGClient) CreateSession(userID string, userEmail string, userIsAdmin bool) (string, error) {
	id := chassis.NewBareID(16)

	_, err := pg.DB.Exec(`INSERT INTO sessions VALUES ($1, $2, $3, $4)`,
		id, userID, userEmail, userIsAdmin)
	if err != nil {
		return "", nil
	}

	return id, nil
}

// LookupSession checks a session token and returns the associated
// user ID, email and admin flag if the session is known.
func (pg *PGClient) LookupSession(token string) (string, string, bool, error) {
	sess := model.Session{}
	err := pg.DB.Get(&sess, lookupSession, token)
	if err == sql.ErrNoRows {
		return "", "", false, ErrSessionNotFound
	}
	if err != nil {
		return "", "", false, err
	}
	return sess.UserID, sess.Email, sess.IsAdmin, nil
}

const lookupSession = `
SELECT token, user_id, email, is_admin
  FROM sessions
 WHERE token = $1`

// UpdateSessions updates all sessions for a user to reflect changes
// in user data.
func (pg *PGClient) UpdateSessions(userID string, userEmail string, userIsAdmin bool) error {
	_, err := pg.DB.Exec(updateSessions, userID, userEmail, userIsAdmin)
	return err
}

const updateSessions = `
UPDATE sessions SET email = $2, is_admin = $3
 WHERE user_id = $1`

// DeleteSession deletes a single session, i.e. logs a user out of
// their current session.
func (pg *PGClient) DeleteSession(token string) error {
	result, err := pg.DB.Exec(`DELETE FROM sessions WHERE token = $1`, token)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ErrSessionNotFound
	}
	return nil
}

// DeleteUserSessions deletes all sessions for a user, i.e. logs the
// user out of all devices where they're logged in.
func (pg *PGClient) DeleteUserSessions(userID string) error {
	_, err := pg.DB.Exec(`DELETE FROM sessions WHERE user_id = $1`, userID)
	if err != nil {
		return err
	}
	return nil
}

// SaveEvent saves an event to the database.
func (pg *PGClient) SaveEvent(label string, eventData interface{}, inTx func() error) error {
	tx, err := pg.DB.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			tx.Rollback()
		}
	}()
	err = chassis.SaveEvent(tx, label, eventData, inTx)
	return err
}

// Parameters to use for random token generation.
const (
	tokenLen     = 6
	dChars       = "0123456789"
	dCharIdxBits = 4                   // 4 bits to represent a character index
	dCharIdxMask = 1<<dCharIdxBits - 1 // All 1-bits, as many as dCharIdxBits
	dCharIdxMax  = 63 / dCharIdxBits   // # of char indices fitting in 63 bits
)

// RandToken generates a random string of digits to use as a login token.
func RandToken() string {
	return chassis.RandString(dChars, dCharIdxBits, dCharIdxMask, dCharIdxMax,
		tokenLen)
}
