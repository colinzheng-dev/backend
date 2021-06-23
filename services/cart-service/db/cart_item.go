package db

import (
	"database/sql"
	//"database/sql"
	//"errors"
	"github.com/jmoiron/sqlx"
	//"strings"

	// Import Postgres DB driver.

	//"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/cart-service/model"
	//"github.com/veganbase/backend/services/cart-service/model/types"
)

func (pg *PGClient) CartItemsByCartId(cartId string) (*[]model.CartItem, error) {
	cartItems := &[]model.CartItem{}
	err := sqlx.Select(pg.DB, cartItems, qCartItemBy + `cart_id = $1`, cartId)
	if err != sql.ErrNoRows && err != nil {
		return nil, err
	}
	return cartItems, nil
}

func (pg *PGClient) CartItemByCartIdAndItemId(cartId string, itemId string) (*model.CartItem, error) {

	item := &model.CartItem{}
	err := sqlx.Get(pg.DB, item, qCartItemBy + `cart_id = $1 and item_id = $2`, cartId, itemId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrCartItemNotFound
		}
		return nil, err
	}
	return item, nil
}

func (pg *PGClient) CartItemByCartIdAndCartItemId(cartId string, cItemId int) (*model.CartItem, error) {

	item := &model.CartItem{}
	err := sqlx.Get(pg.DB, item, qCartItemBy + `cart_id = $1 and id = $2`, cartId, cItemId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrCartItemNotFound
		}
		return nil, err
	}
	return item, nil
}

const qCartItemBy = `
SELECT id, cart_id, item_id, quantity, item_type, other_info, subscribe, delivery_every
  FROM cart_items WHERE `

// CreateCart creates a new cart item.
func (pg *PGClient) CreateCartItem(cart *model.CartItem) error {
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

	rows, err := tx.NamedQuery(qCreateCartItem, cart)
	if err != nil {
		return err
	}
	if rows.Next() {
		err = rows.Scan(&cart.ID)
		if err != nil {
			return err
		}
	}

	return err
}

const qCreateCartItem = `
INSERT INTO
  cart_items (cart_id, item_id, quantity, item_type, other_info, subscribe, delivery_every)
 VALUES (:cart_id, :item_id, :quantity, :item_type, :other_info, :subscribe, :delivery_every)
 ON CONFLICT DO NOTHING
 RETURNING id`

// DeleteCartItems deletes a cart item with a certain id
// this function do not validate the owner of the cart
func (pg *PGClient) DeleteCartItem(id int) error {
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

	// Try to delete the cart item.
	result, err := tx.Exec(qDeleteItem, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ErrCartItemNotFound //TODO: CHANGE ERROR
	}
	return err
}

const qDeleteItem = `
DELETE FROM cart_items WHERE id = $1`


// UpdateCartItem can update the cart item as long as the user is the owner of the cart or the cart has no owner
func (pg *PGClient) UpdateCartItem(item *model.CartItem) error {
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

	//check if cart item exists
	check := &model.CartItem{}
	err = tx.Get(check, qCartItemBy + `id = $1`, item.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrCartItemNotFound
		}
		return err
	}

	result, err := tx.NamedExec(qUpdateCartItem, item)
	if err != nil {
		return err
	}
	_, err = result.RowsAffected()
	if err != nil {
		return err
	}

	return err
}

const qUpdateCartItem = `
UPDATE cart_items
 SET cart_id=:cart_id, item_id=:item_id, 
     quantity=:quantity, other_info=:other_info, 
     subscribe=:subscribe, delivery_every=:delivery_every
 WHERE id = :id`

