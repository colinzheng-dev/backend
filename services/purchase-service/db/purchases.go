package db

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/purchase-service/model"
)

const qPurchaseBy = `
SELECT id, buyer_id, items, delivery_fees, status, site, created_at
FROM purchases WHERE `

// PurchaseById looks up a purchase by its ID.
func (pg *PGClient) PurchaseById(purchaseId string) (*model.Purchase, error) {
	purchase := &model.Purchase{}
	err := sqlx.Get(pg.DB, purchase, qPurchaseBy + `id = $1`, purchaseId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPurchaseNotFound
		}
		return nil, err
	}
	return purchase, nil
}

// PurchasesByOwner retrieves all the purchases of a specific owner
func (pg *PGClient) PurchasesByOwner(owner string, params *chassis.Pagination) (*[]model.Purchase, *uint, error) {
	purchases := &[]model.Purchase{}
	var total uint

	q := qPurchaseBy + `buyer_id = '` + owner + `' ORDER BY created_at DESC `
	if params != nil {
		q += chassis.Paginate(params.Page, params.PerPage)
	}

	if err := sqlx.Select(pg.DB, purchases, q); err != nil && err != sql.ErrNoRows {
		return nil, nil,  err
	}
	if err := chassis.GetTotalResultsFromQuery(q, &total, pg.DB); err != nil {
		return nil, nil, err
	}
	return purchases, &total, nil
}

// CreateCart creates a new purchase along with bookings and orders.
func (pg *PGClient) CreatePurchase(purchase *model.Purchase, orders *[]model.Order, bookings *[]model.Booking) error {
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
	//creating a new purchaseId

	//inserting the purchase
	purchase.Id = chassis.NewPurchaseID()
	rows, err := tx.NamedQuery(qCreatePurchase, purchase)
	if err != nil {
		return err
	}
	if rows.Next() {
		if err = rows.Scan(&purchase.CreatedAt); err != nil {
			return err
		}
		if err = rows.Close(); err != nil {
			return err
		}
	}

	//linking the purchase with orders and inserting them
	for i := range *orders {
		(*orders)[i].Origin = purchase.Id
		(*orders)[i].Id = chassis.NewPurchaseID()
		rows, err := tx.NamedQuery(qCreateOrder, (*orders)[i])
		if err != nil {
			return err
		}
		if rows.Next() {
			if err = rows.Scan(&(*orders)[i].CreatedAt); err != nil {
				return err
			}
			if err = rows.Close(); err != nil {
				return err
			}
		}
	}

	//linking the purchase with bookings and inserting them
	for i := range *bookings {
		(*bookings)[i].Origin = purchase.Id
		(*bookings)[i].Id = chassis.NewPurchaseID()
		rows, err := tx.NamedQuery(qCreateBooking, (*bookings)[i])
		if err != nil {
			return err
		}
		if rows.Next() {
			if err = rows.Scan(&(*bookings)[i].CreatedAt); err != nil {
				return err
			}
			if err = rows.Close(); err != nil {
				return err
			}
		}
	}
	return err
}

const qCreatePurchase = `
INSERT INTO
	purchases (id, buyer_id, items, delivery_fees, status, site)
VALUES (:id, :buyer_id, :items, :delivery_fees, :status, :site)
ON CONFLICT DO NOTHING
RETURNING created_at`

const qCreateOrder = `
INSERT INTO
	orders (id, origin, buyer_id, seller, payment_status, items, delivery_fee, order_info)
VALUES (:id, :origin, :buyer_id, :seller, :payment_status, :items, :delivery_fee, :order_info)
ON CONFLICT DO NOTHING
RETURNING created_at`

const qCreateBooking = `
INSERT INTO
  bookings (id, origin, buyer_id, host, item_id, booking_info, payment_status)
 VALUES ( :id, :origin, :buyer_id, :host, :item_id, :booking_info, :payment_status)
 ON CONFLICT DO NOTHING
 RETURNING created_at`

// UpdatePurchase updates the status of a purchase.
func (pg *PGClient) UpdatePurchase(purchase *model.Purchase) error {
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

	check := &model.Purchase{}
	err = tx.Get(check, qPurchaseBy+`id = $1`, purchase.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrPurchaseNotFound
		}
		return err
	}

	result, err := tx.NamedExec(qUpdatePurchase, purchase)
	if err != nil {
		return err
	}
	_, err = result.RowsAffected()
	if err != nil {
		return err
	}

	return err
}

const qUpdatePurchase = `
UPDATE purchases
SET status = :status
WHERE id = :id`
