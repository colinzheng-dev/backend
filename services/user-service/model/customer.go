package model

import (
	"time"
)

// Customer is a brief reference to a Stripe customer entry.
type Customer struct {
	UserID     string    `json:"user_id" db:"user_id"`
	CustomerID string    `json:"customer_id" db:"customer_id"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}
