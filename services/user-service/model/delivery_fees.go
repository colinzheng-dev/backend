package model

import (
	"time"
)

// DeliveryFees represents delivery fees configuration for a seller.
type DeliveryFees struct {
	ID                string    `json:"id" db:"id"`
	Owner             string    `json:"owner" db:"owner"`
	FreeDeliveryAbove int       `json:"free_delivery_above" db:"free_delivery_above"`
	NormalOrderPrice  int       `json:"normal_order_price" db:"normal_order_price"`
	ChilledOrderPrice int       `json:"chilled_order_price" db:"chilled_order_price"`
	Currency          string    `json:"currency" db:"currency"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
}
