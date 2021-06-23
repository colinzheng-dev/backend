package db

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/cart-service/model"
)

const qCartBy = `
SELECT id, cart_status, owner, created_at
  FROM carts WHERE `

// CartByID looks up a cart by its ID.
func (pg *PGClient) CartByID(cartId string) (*model.Cart, error) {
	cart := &model.Cart{}
	if err := sqlx.Get(pg.DB, cart, qCartBy+`id = $1`, cartId); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrCartNotFound
		}
		return nil, err
	}
	return cart, nil
}


// CartsByOwner retrives all the carts of a specific owner
func (pg *PGClient) CartsByOwner(owner string) (*[]model.Cart, error) {
	carts := &[]model.Cart{}
	if err := sqlx.Select(pg.DB, carts, qCartBy+`owner = $1`, owner); err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	return carts, nil
}


// GetActiveCartByOwner looks up for an active cart of a specific user.
func (pg *PGClient) GetActiveCartByOwner(owner string) (*model.Cart, error) {
	cart := &model.Cart{}
	if err := sqlx.Get(pg.DB, cart, qCartBy+`cart_status = 'active' and owner = $1`, owner); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrCartNotFound
		}
		return nil, err
	}
	return cart, nil
}

// CreateCart creates a new cart.
func (pg *PGClient) CreateCart(cart *model.Cart) (*model.Cart, error) {
	tx, err := pg.DB.Beginx()
	if err != nil {
		return nil, err
	}
	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			tx.Rollback()
		}
	}()

	cartId := chassis.NewID("car")
	cart.ID = cartId

	rows, err := tx.NamedQuery(qCreateCart, cart)
	if err != nil {
		return nil, err
	}
	if rows.Next() {
		err = rows.Scan(&cart.CreatedAt)
		if err != nil {
			return nil, err
		}
	}

	return cart, err
}

const qCreateCart = `
INSERT INTO
  carts (id, cart_status, owner)
 VALUES (:id, :cart_status, :owner)
 ON CONFLICT DO NOTHING
 RETURNING created_at`

// UpdateCart updates the cart status and owner information in the database.
func (pg *PGClient) UpdateCart(cart *model.Cart) error {
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

	check := &model.Cart{}
	err = tx.Get(check, qCartBy+`id = $1`, cart.ID)
	if err == sql.ErrNoRows {
		return ErrCartNotFound
	}
	if err != nil {
		return err
	}

	result, err := tx.NamedExec(qUpdateCart, cart)
	if err != nil {
		return err
	}
	_, err = result.RowsAffected()
	if err != nil {
		return err
	}

	return err
}

const qUpdateCart = `
UPDATE carts
 SET cart_status=:cart_status, owner=:owner
 WHERE id = :id`

// UpdateCart updates the cart status and owner information in the database.
func (pg *PGClient) DeleteCart(cartId string) error {
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

	check := &model.Cart{}
	err = tx.Get(check, qCartBy+`id = $1`, cartId)
	if err == sql.ErrNoRows {
		return ErrCartNotFound
	}
	if err != nil {
		return err
	}

	result, err := tx.Exec(qDeleteCart, check.ID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ErrCartNotFound
	}
	return err
}

const qDeleteCart = `DELETE FROM carts WHERE id = $1 and cart_status = 'not logged in'`
