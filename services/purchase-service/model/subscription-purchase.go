package model

import (
	"github.com/veganbase/backend/chassis"
	"time"
)

type SubscriptionPurchase struct {
	ID          int                      `db:"id" json:"id"`
	Reference   string                   `db:"reference" json:"reference"`
	BuyerID     string                   `db:"buyer_id" json:"buyer_id"`
	AddressID   string                   `db:"address_id" json:"address_id"`
	Status      chassis.ProcessingStatus `db:"status" json:"status"`
	PurchaseID  *string                  `db:"purchase_id" json:"purchase_id"`
	Errors      *string                  `db:"errors" json:"errors"`
	CreatedAt   time.Time                `db:"created_at" json:"created_at"`
	ProcessedAt *time.Time               `db:"processed_at" json:"processed_at"`
}
