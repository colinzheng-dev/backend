package model

import (
	"github.com/veganbase/backend/chassis"
	"time"
)

type SubscriptionPurchaseProcessing struct {
	ID              int                      `db:"id" json:"id"`
	Owner           string                   `db:"owner" json:"owner"`
	Reference       string                   `db:"reference" json:"reference"`
	IsProcessingDay bool                     `db:"is_processing_day" json:"is_processing_day"`
	Status          chassis.ProcessingStatus `db:"status" json:"status"`
	CreatedAt       time.Time                `db:"created_at" json:"created_at"`
	StartedAt       *time.Time               `db:"started_at" json:"started_at"`
	EndedAt         *time.Time                `db:"ended_at" json:"ended_at"`
}
