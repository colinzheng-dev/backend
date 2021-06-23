package db

import (
	"database/sql"

	_ "github.com/lib/pq"

	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/user-service/model"
)

// CreateDeliveryFees creates a new configuration for delivery fees.
func (pg *PGClient) CreateDeliveryFees(delFee *model.DeliveryFees) error {
	tx, err := pg.DB.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	// Generate new delivery fee
	delFee.ID = chassis.NewID("dfe")

	rows, err := tx.NamedQuery(qCreateDeliveryFees, delFee)
	if err != nil {
		return err
	}
	if rows.Next() {
		if err = rows.Scan(&delFee.CreatedAt); err != nil {

			return err
		}
	}

	return err
}

const qCreateDeliveryFees = `
INSERT INTO 
     delivery_fees (id, owner, free_delivery_above, normal_order_price, chilled_order_price, currency)
VALUES (:id, :owner, :free_delivery_above, :normal_order_price, :chilled_order_price, :currency)
RETURNING created_at `

// DeliveryFeesByOwner looks up for delivery fees configurations of a certain owner.
func (pg *PGClient) DeliveryFeesByOwner(id string) (*model.DeliveryFees, error) {
	fees := &model.DeliveryFees{}
	if err := pg.DB.Get(fees, deliveryFeesBy + " owner = $1", id); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrDeliveryFeesNotFound
		}
		return nil, err
	}
	return fees, nil
}

// DeliveryFeesByOwner looks up for delivery fees configurations with a certain id.
func (pg *PGClient) DeliveryFeesById(id string) (*model.DeliveryFees, error) {
	fees := &model.DeliveryFees{}

	if err := pg.DB.Get(fees, deliveryFeesBy + " id = $1", id); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrDeliveryFeesNotFound
		}
		return nil, err
	}
	return fees, nil
}

const deliveryFeesBy = `
SELECT id, owner, free_delivery_above, normal_order_price, chilled_order_price, currency, created_at
FROM delivery_fees
WHERE `

// UpdateDeliveryFees updates delivery fees configuration in the database.
func (pg *PGClient) UpdateDeliveryFees(delFee *model.DeliveryFees) error {
	tx, err := pg.DB.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	check := &model.DeliveryFees{}
	if err = tx.Get(check, deliveryFeesBy + ` owner = $1`, delFee.Owner); err != nil {
		if err == sql.ErrNoRows {
			return ErrDeliveryFeesNotFound
		}
		return err
	}

	// Do the update.
	if _, err = tx.NamedExec(updateDeliveryFees, delFee); err != nil {
		return err
	}

	return nil
}

// id, owner and created_at are read-only
const updateDeliveryFees = `
UPDATE delivery_fees 
SET free_delivery_above = :free_delivery_above, 
    normal_order_price = :normal_order_price, 
    chilled_order_price = :chilled_order_price,
    currency = :currency
WHERE id = :id `

// DeleteDeliveryFees deletes the given delivery fees configuration.
func (pg *PGClient) DeleteDeliveryFees(id string) error {
	result, err := pg.DB.Exec(deleteDeliveryFees, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ErrDeliveryFeesNotFound
	}
	return nil
}

const deleteDeliveryFees = `DELETE FROM delivery_fees WHERE id = $1`
