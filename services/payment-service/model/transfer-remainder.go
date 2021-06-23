package model

import "time"

type TransferRemainder struct {
	TransferId           string    `db:"transfer_id"`           //Stripe's transfer ID
	Destination          string    `db:"destination"`           //user or org ID
	DestinationAccount   string    `db:"destination_account"`   //Destination Stripe Account
	Currency             string    `db:"currency"`              //Currency used on order/booking and delivery fees
	TotalValue           int       `db:"total_value"`           //Total value of the order or booking
	FeeValue             int       `db:"fee_value"`             //How much was charged as fee in %
	TransferredValue     int       `db:"transferred_value"`     //Valued transferred to destination account
	FeeRemainder         float64   `db:"fee_remainder"`         //Remainder that belongs to Veganbase
	TransferredRemainder float64   `db:"transferred_remainder"` //Remainder that must be transferred to destination account
	CreatedAt            time.Time `db:"created_at"`
}
