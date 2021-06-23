package model

import (
	"github.com/veganbase/backend/services/user-service/model/types"
	"time"
)

// PaymentMethod is a brief reference to one Stripe's payment method entry.
type PaymentMethod struct {
	ID              string          `json:"id" db:"id"`
	PaymentMethodID string          `json:"pm_id" db:"pm_id"`
	UserID          string          `json:"user_id" db:"user_id"`
	Description     string          `json:"description,omitempty" db:"description"`
	IsDefault       bool            `json:"is_default" db:"is_default"`
	Type            string          `json:"type" db:"type"`
	OtherInfo       types.OtherInfo `json:"other_info,omitempty" db:"other_info"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
}