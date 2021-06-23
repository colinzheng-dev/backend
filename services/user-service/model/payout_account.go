package model

import (
	"time"
)

// PayoutAccount are the accounts that the payments will be forwarded to each seller or host when a purchase is fulfilled.
type PayoutAccount struct {
	ID            string    `json:"id" db:"id"`
	AccountNumber string    `json:"account" db:"account"`
	Owner         string    `json:"owner" db:"owner"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

type PayoutAccountRequest struct {
	PayoutAccount
	Code string `json:"code,omitempty"`
}
