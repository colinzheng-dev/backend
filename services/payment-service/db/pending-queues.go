package db

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/veganbase/backend/services/payment-service/model"
)

func (pg *PGClient) PendingEvents() (*[]model.PendingEvent, error) {

	pe := &[]model.PendingEvent{}
	if err := sqlx.Select(pg.DB, pe, qPendingEventBy + ` TRUE`); err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	return pe, nil
}

func (pg *PGClient) PendingEventByEventID(eventId string) (*model.PendingEvent, error) {

	pe := &model.PendingEvent{}
	if 	err := sqlx.Get(pg.DB, pe, qPendingEventBy + `event_id = $1`, eventId); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPendingEventNotFound
		}
		return nil, err
	}
	return pe, nil
}

const qPendingEventBy = `
SELECT event_id, intent_id,reason, attempts, last_update, created_at
  FROM pending_events WHERE `

func (pg *PGClient) CreatePendingEvent(pe *model.PendingEvent) error {
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

	rows, err := tx.NamedQuery(qCreatePendingEvent, pe)
	if err != nil {
		return err
	}
	if rows.Next() {
		if err = rows.Scan(&pe.CreatedAt); err != nil {
			return err
		}
		if err = rows.Close(); err != nil {
			return err
		}
	}

	return err
}

const qCreatePendingEvent = `
INSERT INTO
	pending_events ( event_id, intent_id, reason, attempts)
VALUES (:event_id, :intent_id, :reason, :attempts)
ON CONFLICT DO NOTHING
RETURNING created_at`


// DeletePendingEvent deletes a pending event that was already processed
func (pg *PGClient) DeletePendingEvent(eventId string) error {
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

	// Try to delete the cart item.
	result, err := tx.Exec(qDeletePendingEvent, eventId)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ErrPendingEventNotFound
	}
	return err
}

const qDeletePendingEvent = `
DELETE FROM pending_events WHERE event_id = $1`

// UpdatePendingEvent
func (pg *PGClient) UpdatePendingEvent(pe *model.PendingEvent) error {
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

	//check if cart item exists
	check := &model.PendingEvent{}
	err = tx.Get(check, qPendingEventBy + `event_id = $1`, pe.EventID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrPendingEventNotFound
		}
		return err
	}

	result, err := tx.NamedExec(qUpdatePendingEvent, pe)
	if err != nil {
		return err
	}
	_, err = result.RowsAffected()
	if err != nil {
		return err
	}

	return err
}

const qUpdatePendingEvent = `
UPDATE pending_events
SET attempts=:attempts, last_update = now()
WHERE event_id = :event_id`

/***********************************************************************
************************************************************************
************************************************************************/

func (pg *PGClient) PendingTransfers() (*[]model.PendingTransfer, error) {

	pe := &[]model.PendingTransfer{}
	if err := sqlx.Select(pg.DB, pe, qPendingTransferBy + ` TRUE`); err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	return pe, nil
}

//func (pg *PGClient) PendingTransferByID(eventId string) (*model.PendingEvent, error) {
//
//	pe := &model.PendingEvent{}
//	if 	err := sqlx.Get(pg.DB, pe, qPendingTransferBy + `event_id = $1`, eventId); err != nil {
//		if err == sql.ErrNoRows {
//			return nil, ErrPendingTransferNotFound
//		}
//		return nil, err
//	}
//	return pe, nil
//}

const qPendingTransferBy = `
SELECT id, origin, destination, currency, source_transaction, total_value, fee_value, transferred_value, 
       fee_remainder, transferred_remainder, reason, created_at
FROM pending_transfers WHERE `

func (pg *PGClient) CreatePendingTransfer(pt *model.PendingTransfer) error {
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

	rows, err := tx.NamedQuery(qCreatePendingTransfer, pt)
	if err != nil {
		return err
	}
	if rows.Next() {
		if err = rows.Scan(&pt.CreatedAt); err != nil {
			return err
		}
		if err = rows.Close(); err != nil {
			return err
		}
	}

	return err
}

const qCreatePendingTransfer = `
INSERT INTO	pending_transfers (origin, destination, currency, source_transaction,
                               total_value, fee_value, transferred_value, 
                               fee_remainder, transferred_remainder, reason)
VALUES (:origin, :destination, :currency, :source_transaction, :total_value, :fee_value,
        :transferred_value, :fee_remainder, :transferred_remainder, :reason)
ON CONFLICT DO NOTHING
RETURNING created_at`


// DeletePendingTransfer deletes a pending transfer that was already processed
func (pg *PGClient) DeletePendingTransfer(id int64) error {
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

	check := &model.PendingTransfer{}
	if err = tx.Get(check, qPendingTransferBy + `id = $1`, id); err != nil {
		if err == sql.ErrNoRows {
			return ErrPendingTransferNotFound
		}
		return err
	}

	result, err := tx.Exec(qDeletePendingTransfer, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ErrPendingTransferNotFound
	}
	return err
}

const qDeletePendingTransfer = `DELETE FROM pending_transfers WHERE id = $1`
