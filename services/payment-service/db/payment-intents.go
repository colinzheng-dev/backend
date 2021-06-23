package db

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/veganbase/backend/services/payment-service/model"
)

const qPaymentIntentBy = `
SELECT intent_id, origin, status, currency, origin_amount, created_at, last_update
FROM payment_intents WHERE `


// PaymentIntentsByOwner retrieves all the payment intents of a specific owner (purchase_id)
func (pg *PGClient) PaymentIntentsByOwner(owner string) (*[]model.PaymentIntent, error) {
	intents := &[]model.PaymentIntent{}
	err := sqlx.Select(pg.DB, intents, qPaymentIntentBy + `owner = $1`, owner)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	return intents, nil
}

// PaymentIntentsByOwner retrieves all the payment intents of a specific owner (purchase_id)
func (pg *PGClient) SuccessfulPaymentIntentsByOrigin(origin string) (*[]model.PaymentIntent, error) {
	intents := &[]model.PaymentIntent{}
	err := sqlx.Select(pg.DB, intents, qPaymentIntentBy + `origin = $1`, origin)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	return intents, nil
}

// PaymentIntentByIntentId retrieves a  payment intents by it's stripe's id
func (pg *PGClient) PaymentIntentByIntentId(id string) (*model.PaymentIntent, error) {
	intent := &model.PaymentIntent{}
	if err := sqlx.Get(pg.DB, intent, qPaymentIntentBy+` intent_id = $1`, id); err != nil  {
		if err == sql.ErrNoRows {
			return nil, ErrPaymentIntentNotFound
		}
		return nil, err
	}
	return intent, nil
}


// CreatePaymentIntent creates a new payment intent entry to save on database a
func (pg *PGClient) CreatePaymentIntent(pi *model.PaymentIntent) error {
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

	rows, err := tx.NamedQuery(qCreatePaymentIntent, pi)
	if err != nil {
		return err
	}
	if rows.Next() {
		if err = rows.Scan(&pi.CreatedAt); err != nil {
			return err
		}
		if err = rows.Close(); err != nil {
			return err
		}
	}

	return err
}

const qCreatePaymentIntent = `
INSERT INTO
	payment_intents ( intent_id, origin, status, currency, origin_amount)
VALUES (:intent_id, :origin, :status, :currency, :origin_amount)
ON CONFLICT DO NOTHING
RETURNING created_at`

// UpdatePaymentIntent updates the status and last_update fields of a payment intent.
func (pg *PGClient) UpdatePaymentIntent(pi *model.PaymentIntent) error {
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

	check := &model.PaymentIntent{}
	err = tx.Get(check, qPaymentIntentBy+`intent_id = $1`, pi.StripeIntentId)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrPaymentIntentNotFound
		}
		return err
	}

	result, err := tx.NamedExec(qUpdatePaymentIntent, pi)
	if err != nil {
		return err
	}
	_, err = result.RowsAffected()
	if err != nil {
		return err
	}

	return err
}

const qUpdatePaymentIntent = `
UPDATE payment_intents
SET status = :status, last_update = now()
WHERE intent_id = :intent_id
RETURNING last_update`
