package db

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/veganbase/backend/services/purchase-service/model"
)

const qSubscriptionPurchaseBy = `
SELECT id, reference, buyer_id, address_id, status, purchase_id, errors, created_at, processed_at
FROM subscription_purchases WHERE `

// SubscriptionPurchaseByID looks up a purchase by its ID.
func (pg *PGClient) SubscriptionPurchaseByID(purchaseId string) (*model.SubscriptionPurchase, error) {
	purchase := &model.SubscriptionPurchase{}
	err := sqlx.Get(pg.DB, purchase, qSubscriptionPurchaseBy + `id = $1`, purchaseId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPurchaseNotFound
		}
		return nil, err
	}
	return purchase, nil
}

// SubscriptionPurchasesByReferenceAndStatus retrieves all subscription purchases of a certain reference date
// and with an optional status
func (pg *PGClient) SubscriptionPurchasesByReferenceAndStatus(ref string, status *string) (*[]model.SubscriptionPurchase, error) {
	purchases := &[]model.SubscriptionPurchase{}
	q := qSubscriptionPurchaseBy + `reference = '` + ref + `' `
	if status != nil {
		q += `AND status ='` + *status + `'`
	}
	if err := sqlx.Select(pg.DB, purchases, q); err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	return purchases, nil
}

// CreateCart creates a new purchase along with bookings and orders.
func (pg *PGClient) CreateSubscriptionPurchases(ref string) error {
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

	//inserting the purchase
	_, err = tx.Exec(qCreateSubscriptionPurchases, ref)
	if err != nil {
		return err
	}

	if err = updateNextDeliveries(tx, ref); err != nil {
		return err
	}

	return nil
}

const qCreateSubscriptionPurchases = `
INSERT INTO
	subscription_purchases (reference, buyer_id, address_id)
SELECT DISTINCT to_date($1,'YYYY/MM/DD'), owner, address_id 
		FROM subscription_items 
		WHERE status = 'active' AND EXTRACT(MONTH FROM to_date($1,'YYYY/MM/DD')) = next_delivery
ON CONFLICT DO NOTHING`

// UpdatePurchase updates the status of a purchase.
func (pg *PGClient) UpdateSubscriptionPurchase(subs *model.SubscriptionPurchase) error {
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

	check := &model.SubscriptionPurchase{}
	if err = tx.Get(check, qSubscriptionPurchaseBy + `id = $1`, subs.ID); err != nil {
		if err == sql.ErrNoRows {
			return ErrPurchaseNotFound
		}
		return err
	}

	result, err := tx.NamedExec(qUpdateSubscriptionPurchase, subs)
	if err != nil {
		return err
	}
	_, err = result.RowsAffected()
	if err != nil {
		return err
	}

	return err
}

const qUpdateSubscriptionPurchase = `
UPDATE subscription_purchases
SET status = :status, 
    processed_at = :processed_at, 
    purchase_id = :purchase_id,
	errors = :errors
WHERE id = :id`



func updateNextDeliveries(tx *sqlx.Tx, ref string) error {
	if _, err := tx.Exec(qUpdateNextDeliveries, ref); err != nil {
		return err
	}
	return nil
}

const qUpdateNextDeliveries = `
UPDATE subscription_items
SET last_delivery = now(),
	next_delivery = mod(next_delivery + delivery_every, 12)
WHERE status = 'active' AND EXTRACT(MONTH FROM to_date($1,'YYYY/MM/DD')) = next_delivery `
