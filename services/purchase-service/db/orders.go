package db

import (
	"database/sql"
	"github.com/veganbase/backend/chassis"

	//"database/sql"
	//"errors"
	"github.com/jmoiron/sqlx"
	//"strings"

	// Import Postgres DB driver.

	//"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/purchase-service/model"
	//"github.com/veganbase/backend/services/cart-service/model/types"
)

// OrdersById retrieves a specific order.
func (pg *PGClient) OrderById(orderId string) (*model.Order, error) {
	order := &model.Order{}
	if err := sqlx.Get(pg.DB, order, qOrderBy + `id = $1`, orderId); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	return order, nil
}

// OrdersBySeller retrieves all the orders of a specific seller
func (pg *PGClient) OrdersBySeller(sellerId string, params *chassis.Pagination) (*[]model.Order, *uint, error) {
	orders := &[]model.Order{}
	var total uint

	q := qOrderBy+ `seller = '` + sellerId + `' ORDER BY created_at DESC `
	if params != nil {
		q += chassis.Paginate(params.Page, params.PerPage)
	}

	if err := sqlx.Select(pg.DB, orders, q); err != nil && err != sql.ErrNoRows {
		return nil, nil,  err
	}
	if err := chassis.GetTotalResultsFromQuery(q, &total, pg.DB); err != nil {
		return nil, nil, err
	}
	return orders, &total, nil
}

// OrdersByPurchase retrieves all the orders originated by a specific purchased
func (pg *PGClient) OrdersByPurchase(purchaseId string, params *chassis.Pagination) (*[]model.Order, *uint, error) {
	orders := &[]model.Order{}
	var total uint
	q := qOrderBy+ `origin = '` + purchaseId + `' ORDER BY created_at DESC `
	if params != nil {
		q += chassis.Paginate(params.Page, params.PerPage)
	}

	if err := sqlx.Select(pg.DB, orders, q); err != nil && err != sql.ErrNoRows {
		return nil, nil,  err
	}
	if err := chassis.GetTotalResultsFromQuery(q, &total, pg.DB); err != nil {
		return nil, nil, err
	}
	return orders, &total, nil
}


// OrdersByPurchase retrieves all the orders originated by a specific purchased
func (pg *PGClient) OrdersByBuyer(buyerId string, params *chassis.Pagination) (*[]model.Order, *uint, error) {
	orders := &[]model.Order{}
	var total uint
	q := qOrderBy+ `buyer_id = '` + buyerId + `' ORDER BY created_at DESC `
	if params != nil {
		q += chassis.Paginate(params.Page, params.PerPage)
	}

	if err := sqlx.Select(pg.DB, orders, q); err != nil && err != sql.ErrNoRows {
		return nil, nil,  err
	}
	if err := chassis.GetTotalResultsFromQuery(q, &total, pg.DB); err != nil {
		return nil, nil, err
	}
	return orders, &total, nil
}


const qOrderBy = `
SELECT id, origin, buyer_id, seller, payment_status, items, delivery_fee, order_info, created_at
FROM orders WHERE `


// UpdateOrder updates the payment_status of an order.
func (pg *PGClient) UpdateOrder(order *model.Order) error {
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

	//check if order exists
	check := &model.Order{}
	err = tx.Get(check, qOrderBy + `id = $1`, order.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrOrderNotFound
		}
		return err
	}

	result, err := tx.NamedExec(qUpdateOrder, order)
	if err != nil {
		return err
	}
	_, err = result.RowsAffected()
	if err != nil {
		return err
	}

	return err
}

const qUpdateOrder = `
UPDATE orders
SET payment_status=:payment_status, 
    order_info=:order_info
WHERE id = :id `

