package model

import (
	"time"
)

// Transfer is the base model for record transfers.
type Transfer struct {
	TransferId         string    `db:"transfer_id"`         //Stripe's transfer ID
	Origin             string    `db:"origin"`              //Purchase id that originates this transfer
	Destination        string    `db:"destination"`         //User or org ID
	DestinationAccount string    `db:"destination_account"` //Destination Stripe Account
	Currency           string    `db:"currency"`            //Currency used on order/booking and delivery fees
	Amount             int64     `db:"amount"`              //Valued transferred to destination account
	CreatedAt          time.Time `db:"created_at"`
}
