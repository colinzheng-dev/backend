package db

import "github.com/veganbase/backend/services/payment-service/model"

func (pg *PGClient) CreateTransferRemainder(remainder *model.TransferRemainder) error {
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

	rows, err := tx.NamedQuery(qCreateTransferRemainder, remainder)
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

const qCreateTransferRemainder = `
INSERT INTO
	transfer_remainders ( transfer_id, destination,currency, destination_account, total_value, transferred_value, fee_value, fee_remainder, transferred_remainder)
VALUES (:transfer_id, :destination, :currency, :destination_account, :total_value, :transferred_value, :fee_value, :fee_remainder, :transferred_remainder)
ON CONFLICT DO NOTHING
RETURNING created_at`

