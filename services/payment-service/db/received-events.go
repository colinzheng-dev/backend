package db

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/veganbase/backend/services/payment-service/model"
)
const qReceivedEventBy = `
SELECT event_id, idempotency_key, event_type, is_handled, created_at
FROM received_events 
WHERE `

// ReceivedEventByEventId
func (pg *PGClient) ReceivedEventByEventId(id string) (*model.ReceivedEvent, error) {
	event := &model.ReceivedEvent{}
	err := sqlx.Get(pg.DB, event, qReceivedEventBy + `event_id = $1`, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrReceivedEventNotFound
		}
		return nil, err
	}
	return event, nil
}


func (pg *PGClient) CreateReceivedEvent(event *model.ReceivedEvent) error {
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

	rows, err := tx.NamedQuery(qCreateReceivedEvent, event)
	if err != nil {
		return err
	}
	if rows.Next() {
		if err = rows.Scan(&event.CreatedAt); err != nil {
			return err
		}
		if err = rows.Close(); err != nil {
			return err
		}
	}

	return err
}

const qCreateReceivedEvent = `
INSERT INTO
	received_events ( event_id, idempotency_key, event_type, is_handled)
VALUES (:event_id, :idempotency_key, :event_type, :is_handled)
ON CONFLICT DO NOTHING
RETURNING created_at`



// UpdateReceivedEvent updates the status of a event.
func (pg *PGClient) UpdateReceivedEvent(event *model.ReceivedEvent) error {
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

	check := &model.ReceivedEvent{}
	err = tx.Get(check, qReceivedEventBy + `event_id = $1`, event.EventId)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrReceivedEventNotFound
		}
		return err
	}

	result, err := tx.NamedExec(qUpdateReceivedEvent, event)
	if err != nil {
		return err
	}
	_, err = result.RowsAffected()
	if err != nil {
		return err
	}

	return err
}

const qUpdateReceivedEvent = `
UPDATE received_events
SET is_handled = :is_handled
WHERE event_id = :event_id `
