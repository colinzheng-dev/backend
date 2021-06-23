package db

import "github.com/veganbase/backend/services/payment-service/model"

func (pg *PGClient) CreateErrorLog(log *model.ErrorLog) error {
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

	rows, err := tx.NamedQuery(qCreateErrorLog, log)
	if err != nil {
		return err
	}
	if rows.Next() {
		if err = rows.Scan(&log.CreatedAt); err != nil {
			return err
		}
		if err = rows.Close(); err != nil {
			return err
		}
	}

	return err
}

const qCreateErrorLog = `
INSERT INTO
	error_logs ( event_id, error )
VALUES (:event_id, :error)
ON CONFLICT DO NOTHING
RETURNING created_at`