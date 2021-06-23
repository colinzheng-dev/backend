package db

import (
	"database/sql"

	_ "github.com/lib/pq"

	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/user-service/model"
)

// CreatePaymentMethod creates a new payment method.
func (pg *PGClient) CreatePaymentMethod(pmt *model.PaymentMethod) error {
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

	// Generate new payment method ID
	pmt.ID = chassis.NewID("pmt")

	rows, err := tx.NamedQuery(qCreatePaymentMethod, pmt)
	if err != nil {
		return err
	}
	if rows.Next() {
		if err = rows.Scan(&pmt.CreatedAt); err != nil {

			return err
		}
	}

	return err
}

const qCreatePaymentMethod = `
INSERT INTO
  payment_methods (id, pm_id, user_id, description, is_default, type, other_info)
 VALUES (:id, :pm_id, :user_id, :description, :is_default, :type, :other_info)
 RETURNING created_at`

// PaymentMethodsByUserId looks up for all payment methods of a certain user.
func (pg *PGClient) PaymentMethodsByUserId(id string) (*[]model.PaymentMethod, error) {
	pmts := &[]model.PaymentMethod{}
	if err := pg.DB.Select(pmts, paymentMethodBy + "user_id = $1", id); err != nil && err != sql.ErrNoRows  {
		return nil, err
	}
	return pmts, nil
}


// DefaultPaymentMethodByUserId looks the default payment methods of a certain user.
func (pg *PGClient) DefaultPaymentMethodByUserId(id string) (*model.PaymentMethod, error) {
	pmt := &model.PaymentMethod{}
	if err := pg.DB.Get(pmt, paymentMethodBy + "user_id = $1 and is_default = true", id); err != nil  {
		if err == sql.ErrNoRows {
			return nil, ErrPaymentMethodNotFound
		}
		return nil, err
	}
	return pmt, nil
}


// PaymentMethodById looks up for a specific payment method.
func (pg *PGClient) PaymentMethodById(id string) (*model.PaymentMethod, error) {
	acc := &model.PaymentMethod{}

	if err := pg.DB.Get(acc, paymentMethodBy+"id = $1", id); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPaymentMethodNotFound
		}
		return nil, err
	}
	return acc, nil
}

const paymentMethodBy = `
SELECT id, pm_id, user_id, description, is_default, type, other_info, created_at
  FROM payment_methods
 WHERE `

// UpdatePayoutAccount updates a payout account in the database.
func (pg *PGClient) UpdatePaymentMethod(pmt *model.PaymentMethod) error {
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

	check := &model.PaymentMethod{}
	if err = tx.Get(check, paymentMethodBy+`id = $1`, pmt.ID); err != nil {
		if err == sql.ErrNoRows {
			return ErrPaymentMethodNotFound
		}
		return err
	}

	// Do the update.
	if _, err = tx.NamedExec(updatePaymentMethod, pmt); err != nil {
		return err
	}

	return nil
}

const updatePaymentMethod = `
UPDATE payment_methods 
SET pm_id=:pm_id, is_default = :is_default, type = :type, other_info = :other_info
WHERE id = :id`

// DeletePaymentMethod deletes the given payout account.
func (pg *PGClient) DeletePaymentMethod(id string) error {
	result, err := pg.DB.Exec(deletePaymentMethod, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ErrPaymentMethodNotFound
	}
	return nil
}

const deletePaymentMethod = "DELETE FROM payment_methods WHERE id = $1"
