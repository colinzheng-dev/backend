package db

import (
	"database/sql"

	_ "github.com/lib/pq"

	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/user-service/model"
)

// CreatePayoutAccount creates a new payout account.
func (pg *PGClient) CreatePayoutAccount(acc *model.PayoutAccount) error {
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

	// Generate new payout account ID
	acc.ID = chassis.NewID("acc")

	rows, err := tx.NamedQuery(qCreatePayoutAccount, acc)
	if err != nil {
		return err
	}
	if rows.Next() {
		if err = rows.Scan(&acc.CreatedAt); err != nil {

			return err
		}
	}

	return err
}

const qCreatePayoutAccount = `
INSERT INTO
  payout_accounts (id, account, owner)
 VALUES (:id, :account, :owner)
 RETURNING created_at`

// PayoutAccountByOwner looks up for a payout account of a certain owner.
func (pg *PGClient) PayoutAccountByOwner(id string) (*model.PayoutAccount, error) {
	acc := &model.PayoutAccount{}
	if err := pg.DB.Get(acc, payoutAccountBy + "owner = $1", id); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPayoutAccountNotFound
		}
		return nil, err
	}
	return acc, nil
}

// PayoutAccountById looks up for a specific payout account.
func (pg *PGClient) PayoutAccountById(id string) (*model.PayoutAccount, error) {
	acc := &model.PayoutAccount{}

	if err := pg.DB.Get(acc, payoutAccountBy+"id = $1", id); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPayoutAccountNotFound
		}
		return nil, err
	}
	return acc, nil
}

const payoutAccountBy = `
SELECT id, account, owner, created_at
  FROM payout_accounts
 WHERE `

// UpdatePayoutAccount updates a payout account in the database.
func (pg *PGClient) UpdatePayoutAccount(acc *model.PayoutAccount) error {
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

	check := &model.PayoutAccount{}
	if err = tx.Get(check, payoutAccountBy+`owner = $1`, acc.Owner); err != nil {
		if err == sql.ErrNoRows {
			return ErrPayoutAccountNotFound
		}
		return err
	}

	// Do the update.
	if _, err = tx.NamedExec(updatePayoutAccount, acc); err != nil {
		return err
	}

	return nil
}

// id, and created_at are read-only
const updatePayoutAccount = `
UPDATE payout_accounts 
SET account=:account, owner =:owner
WHERE id = :id`

// DeletePayoutAccount deletes the given payout account.
// TODO: ADD SOME SORT OF ARCHIVAL MECHANISM INSTEAD.
func (pg *PGClient) DeletePayoutAccount(id string) error {
	result, err := pg.DB.Exec(deletePayoutAccount, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ErrPayoutAccountNotFound
	}
	return nil
}

const deletePayoutAccount = "DELETE FROM payout_accounts WHERE id = $1"
