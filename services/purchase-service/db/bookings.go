package db

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/purchase-service/model"
)

// BookingById retrieves a specific booking.
func (pg *PGClient) BookingById(bookingId string) (*model.Booking, error) {
	booking := &model.Booking{}
	if err := sqlx.Get(pg.DB, booking, qBookingBy + `id = $1`, bookingId); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrBookingNotFound
		}
		return nil, err
	}
	return booking, nil
}

// BookingsByHost retrieves all the bookings of a specific host
func (pg *PGClient) BookingsByHost(hostId string, params *chassis.Pagination) (*[]model.Booking, *uint, error) {
	bookings := &[]model.Booking{}
	var total uint
	q := qBookingBy+ `host = '` + hostId + `' ORDER BY created_at DESC `
	if params != nil {
		q += chassis.Paginate(params.Page, params.PerPage)
	}

	if err := sqlx.Select(pg.DB, bookings, q); err != nil && err != sql.ErrNoRows {
		return nil, nil,  err
	}
	if err := chassis.GetTotalResultsFromQuery(q, &total, pg.DB); err != nil {
		return nil, nil, err
	}
	return bookings, &total, nil
}

// BookingsByPurchase retrieves all bookings originated by a specific purchased
func (pg *PGClient) BookingsByPurchase(purchaseId string, params *chassis.Pagination) (*[]model.Booking, *uint, error) {
	bookings := &[]model.Booking{}
	var total uint
	q := qBookingBy+ ` origin = '` + purchaseId + `' ORDER BY created_at DESC `
	if params != nil {
		q += chassis.Paginate(params.Page, params.PerPage)
	}
	if err := sqlx.Select(pg.DB, bookings, q); err != nil && err != sql.ErrNoRows {
		return nil, nil,  err
	}
	if err := chassis.GetTotalResultsFromQuery(q, &total, pg.DB); err != nil {
		return nil, nil, err
	}
	return bookings, &total, nil
}


// BookingsByBuyer retrieves all the bookings of a specific buyer
func (pg *PGClient) BookingsByBuyer(buyerId string, params *chassis.Pagination) (*[]model.Booking, *uint, error) {
	bookings := &[]model.Booking{}
	var total uint
	q := qBookingBy+ ` buyer_id = '` + buyerId + `' ORDER BY created_at DESC `
	if params != nil {
		q += chassis.Paginate(params.Page, params.PerPage)
	}

	if err := sqlx.Select(pg.DB, bookings, q); err != nil && err != sql.ErrNoRows {
		return nil, nil,  err
	}
	if err := chassis.GetTotalResultsFromQuery(q, &total, pg.DB); err != nil {
		return nil, nil, err
	}
	return bookings, &total, nil
}


const qBookingBy = `
SELECT id, origin, buyer_id, host, item_id, booking_info, payment_status, created_at
FROM bookings WHERE `


// UpdateBooking updates the payment_status of a booking
func (pg *PGClient) UpdateBooking(booking *model.Booking) error {
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

	//check if booking exists
	check := &model.Booking{}
	err = tx.Get(check, qBookingBy + `id = $1`, booking.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrOrderNotFound
		}
		return err
	}

	result, err := tx.NamedExec(qUpdateBooking, booking)
	if err != nil {
		return err
	}
	_, err = result.RowsAffected()
	if err != nil {
		return err
	}

	return err
}

const qUpdateBooking = `
UPDATE bookings
SET payment_status=:payment_status,
    booking_info =:booking_info
WHERE id = :id `

