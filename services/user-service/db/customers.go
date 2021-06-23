package db

import (
	"database/sql"

	_ "github.com/lib/pq"

	"github.com/veganbase/backend/services/user-service/model"
)

// CreateCustomer creates a new customer.
func (pg *PGClient) CreateCustomer(cus *model.Customer) error {
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

	rows, err := tx.NamedQuery(qCreateCustomer, cus)
	if err != nil {
		return err
	}
	if rows.Next() {
		if err = rows.Scan(&cus.CreatedAt); err != nil {

			return err
		}
	}

	return err
}

const qCreateCustomer = `
INSERT INTO
  customers (user_id, customer_id)
 VALUES (:user_id, :customer_id)
 RETURNING created_at`


// CustomerByUserId looks for the customer entry of a specific user.
func (pg *PGClient) CustomerByUserId(id string) (*model.Customer, error) {
	cus := &model.Customer{}
	if err := pg.DB.Get(cus, customerBy + "user_id = $1", id); err != nil  {
		if err == sql.ErrNoRows {
			return nil, ErrCustomerNotFound
		}
		return nil, err
	}
	return cus, nil
}



const customerBy = `
SELECT user_id, customer_id, created_at
  FROM customers
 WHERE `


// DeleteCustomer deletes the customer entry of a given user.
func (pg *PGClient) DeleteCustomer(id string) error {
	result, err := pg.DB.Exec(deleteCustomer, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ErrCustomerNotFound
	}
	return nil
}

const deleteCustomer = "DELETE FROM customers WHERE user_id = $1"
