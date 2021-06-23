package db

import (
	"database/sql"
	"fmt"
	"github.com/veganbase/backend/services/webhook-service/model"
	"strings"
)

const (
	BackoffBase = 5
)

func (pg *PGClient) EventByID(eventID string) (*model.Event, error) {
	ev := model.Event{}
	if err := pg.DB.Get(&ev, qGetEventBy+` event_id = $1 `, eventID); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrEventNotFound
		}
		return nil, err
	}
	return &ev, nil
}

func (pg *PGClient) EventsByDestination(destination string) (*[]model.Event, error) {
	events := []model.Event{}
	if err := pg.DB.Select(&events, qGetEventBy+` destination = $1`, destination); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrEventNotFound
		}
		return nil, err
	}
	return &events, nil
}

func (pg *PGClient) PendingEvents() (*[]model.Event, error) {
	events := []model.Event{}
	q := qGetEventBy + ` retry = true AND attempts > 0 AND (now() > backoff_until OR backoff_until IS NULL)`
	if err := pg.DB.Select(&events, q); err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	return &events, nil
}

const qGetEventBy = `
	SELECT 
		event_id, destination, type, livemode, payload, created_at, sent, sent_at, 
		attempts, retry, last_retry, backoff_until, received_at
	FROM events
	WHERE 
`

// CreateWebhook creates a new webhook entry.
func (pg *PGClient) AddEvent(e *model.Event) error {
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

	rows, err := tx.NamedQuery(qAddEvent, e)
	if err != nil {
		return err
	}
	if rows.Next() {
		err = rows.Scan(&e.ReceivedAt)
		if err != nil {
			return err
		}
	}

	return err
}

const qAddEvent = `
INSERT INTO events (event_id, destination, type, livemode, payload, created_at)
VALUES (:event_id, :destination, :type, :livemode, :payload, :created_at)
ON CONFLICT DO NOTHING
RETURNING received_at`

func (pg *PGClient) UpdateEvent(ev *model.Event) error {
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

	check := &model.Event{}

	if err = tx.Get(check, qGetEventBy+` event_id = $1`, ev.EventID); err != nil {
		if err == sql.ErrNoRows {
			return ErrEventNotFound
		}
		return err
	}

	result, err := tx.NamedExec(qUpdateEvent, ev)
	if err != nil {
		return err
	}
	_, err = result.RowsAffected()
	if err != nil {
		return err
	}

	return err
}

const qUpdateEvent = `
UPDATE events
SET sent=:sent, sent_at =:sent_at, attempts=:attempts, retry=:retry, 
    last_retry=:last_retry, backoff_until=:backoff_until
WHERE event_id = :event_id `

func (pg *PGClient) SetSentStatus(eventID string) error {
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

	check := &model.Event{}

	if err = tx.Get(check, qGetEventBy+` event_id = $1`, eventID); err != nil {
		if err == sql.ErrNoRows {
			return ErrEventNotFound
		}
		return err
	}
	check.Sent = true
	check.Attempts++
	check.Retry = false

	result, err := tx.NamedExec(qSetSentStatus, check)
	if err != nil {
		return err
	}
	_, err = result.RowsAffected()
	if err != nil {
		return err
	}

	return err
}

const qSetSentStatus = `
UPDATE events
SET sent=:sent, attempts=:attempts, retry=:retry, sent_at =now()
WHERE event_id = :event_id `

func (pg *PGClient) IncreaseFailedAttempts(eventID string) error {
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

	check := &model.Event{}
	if err = tx.Get(check, qGetEventBy+` event_id = $1`, eventID); err != nil {
		if err == sql.ErrNoRows {
			return ErrEventNotFound
		}
		return err
	}
	check.Attempts++
	backoff := fmt.Sprintf(`now() + INTERVAL '%d min'`, check.Attempts * BackoffBase)
	q := strings.ReplaceAll(qIncreaseAttempts, "<backoff_until>", backoff )
	result, err := tx.Exec(q, check.Attempts, check.EventID)
	if err != nil {
		return err
	}
	_, err = result.RowsAffected()
	if err != nil {
		return err
	}

	return err
}

const qIncreaseAttempts = `
UPDATE events
SET attempts=$1, backoff_until=<backoff_until>, last_retry = now()
WHERE event_id = $2 `

func (pg *PGClient) DisableRetryFlag(eventID string) error {
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

	check := &model.Event{}
	if err = tx.Get(check, qGetEventBy+` event_id = $1`, eventID); err != nil {
		if err == sql.ErrNoRows {
			return ErrEventNotFound
		}
		return err
	}
	check.Retry = false

	result, err := tx.NamedExec(qDisableRetry, check)
	if err != nil {
		return err
	}
	_, err = result.RowsAffected()
	if err != nil {
		return err
	}

	return err
}

const qDisableRetry = `
UPDATE events
SET retry=:retry
WHERE event_id = :event_id `
