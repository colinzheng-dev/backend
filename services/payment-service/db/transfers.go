package db

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/veganbase/backend/services/payment-service/model"
	"time"
)

const qTransferBy = `
SELECT transfer_id, origin, destination, destination_account, currency, amount, created_at
FROM transfers WHERE `

// TransfersByDestination retrieves all transfers made to a specific account.
func (pg *PGClient) TransfersByDestination(destination string) (*[]model.Transfer, error) {
	intents := &[]model.Transfer{}
	err := sqlx.Select(pg.DB, intents, qTransferBy+`owner = $1`, destination)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	return intents, nil
}

// TransfersByOrigin retrieves all transfers made to fulfill a specific origin (order/booking).
func (pg *PGClient) TransfersByOrigin(origin string) (*[]model.Transfer, error) {
	intents := &[]model.Transfer{}
	err := sqlx.Select(pg.DB, intents, qTransferBy+`origin = $1`, origin)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	return intents, nil
}

// CreatePaymentIntent creates a new payment intent entry to save on database a
func (pg *PGClient) CreateTransfer(tr *model.Transfer) error {
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

	rows, err := tx.NamedQuery(qCreateTransfer, tr)
	if err != nil {
		return err
	}
	if rows.Next() {
		if err = rows.Scan(&tr.CreatedAt); err != nil {
			return err
		}
		if err = rows.Close(); err != nil {
			return err
		}
	}

	return err
}

const qCreateTransfer = `
INSERT INTO
	transfers ( transfer_id, origin, destination, destination_account, currency, amount)
VALUES (:transfer_id, :origin, :destination, :destination_account, :currency, :amount)
ON CONFLICT DO NOTHING
RETURNING created_at`

func (pg *PGClient) CreateTransfers(remainder *model.TransferRemainder, origin string) error {
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

	transfer := model.Transfer{
		Origin:             origin,
		DestinationAccount: remainder.DestinationAccount,
		Destination:        remainder.Destination,
		Currency:           remainder.Currency,
		Amount:             int64(remainder.TransferredValue),
		CreatedAt:          time.Time{},
	}
	rows, err := tx.NamedQuery(qCreateTransfer, transfer)
	if err != nil {
		return err
	}
	if rows.Next() {
		if err = rows.Scan(&transfer.CreatedAt); err != nil {
			return err
		}
		if err = rows.Close(); err != nil {
			return err
		}
	}

	rows, err = tx.NamedQuery(qCreateTransferRemainder, remainder)
	if err != nil {
		return err
	}
	if rows.Next() {
		if err = rows.Scan(&remainder.CreatedAt); err != nil {
			return err
		}
		if err = rows.Close(); err != nil {
			return err
		}
	}

	return err
}
