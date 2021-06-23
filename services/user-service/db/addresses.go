package db

import (
	"database/sql"
	"github.com/jmoiron/sqlx"

	_ "github.com/lib/pq"

	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/user-service/model"
)

// CreateAddress creates a new address.
func (pg *PGClient) CreateAddress(addr *model.Address) error {
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

	// Generate new payment address ID
	addr.ID = chassis.NewID("adr")

	if addr.IsDefault {
		if err = turnIsDefaultOff(tx, addr.Owner); err != nil {
			return err
		}
	}

	rows, err := tx.NamedQuery(qCreateAddress, addr)
	if err != nil {
		return err
	}
	if rows.Next() {
		if err = rows.Scan(&addr.CreatedAt); err != nil {

			return err
		}
	}

	return err
}

const qCreateAddress = `
INSERT INTO
  addresses (id, owner, description, street_address, city, postcode, country, 
             region_postal, house_number, is_default, recipient, coordinates)
 VALUES (:id, :owner, :description, :street_address, :city, :postcode, :country, 
         :region_postal, :house_number, :is_default, :recipient, :coordinates)
 RETURNING created_at `

// AddressesByUserId looks up for all addresses of a certain user.
func (pg *PGClient) AddressesByUserId(id string) (*[]model.Address, error) {
	addrs := &[]model.Address{}
	if err := pg.DB.Select(addrs, qAddressBy + "owner = $1", id); err != nil && err != sql.ErrNoRows  {
		return nil, err
	}
	return addrs, nil
}


// DefaultAddressByUserId looks the default address of a certain user.
func (pg *PGClient) DefaultAddressByUserId(id string) (*model.Address, error) {
	addr := &model.Address{}
	if err := pg.DB.Get(addr, qAddressBy + "owner = $1 and is_default = true", id); err != nil  {
		if err == sql.ErrNoRows {
			return nil, ErrAddressNotFound
		}
		return nil, err
	}
	return addr, nil
}


// AddressById looks up for a specific address.
func (pg *PGClient) AddressById(id string) (*model.Address, error) {
	addr := &model.Address{}

	if err := pg.DB.Get(addr, qAddressBy + "id = $1", id); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrAddressNotFound
		}
		return nil, err
	}
	return addr, nil
}

const qAddressBy = `
SELECT id, owner, description, street_address, city, postcode, region_postal, country, house_number, 
       is_default, recipient, coordinates, created_at
  FROM addresses
 WHERE `

// UpdateAddress updates an address in the database.
func (pg *PGClient) UpdateAddress(addr *model.Address) error {
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

	check := &model.Address{}
	if err = tx.Get(check, qAddressBy + ` id = $1`, addr.ID); err != nil {
		if err == sql.ErrNoRows {
			return ErrAddressNotFound
		}
		return err
	}

	if check.IsDefault != addr.IsDefault && addr.IsDefault {
		if err = turnIsDefaultOff(tx, addr.Owner); err != nil {
			return err
		}
	}

	// Do the update.
	if _, err = tx.NamedExec(qUpdateAddress, addr); err != nil {
		return err
	}

	return nil
}

const qUpdateAddress = `
UPDATE addresses 
SET description=:description, street_address = :street_address, 
    city = :city, postcode = :postcode, region_postal = :region_postal, country = :country,
    house_number = :house_number, is_default = :is_default, recipient=:recipient, coordinates=:coordinates
WHERE id = :id`


func turnIsDefaultOff(tx *sqlx.Tx, owner string )error {
	// Do the update.
	if _, err := tx.Exec(qTurnDefaultOff, owner); err != nil {
		return err
	}
	return nil
}

const qTurnDefaultOff = `
UPDATE addresses 
SET is_default = false WHERE owner = $1`

// DeleteAddress deletes the given address.
func (pg *PGClient) DeleteAddress(id string) error {
	result, err := pg.DB.Exec(qDeleteAddress, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ErrAddressNotFound
	}
	return nil
}

const qDeleteAddress = "DELETE FROM addresses WHERE id = $1"
