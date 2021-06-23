package model

import (
	"encoding/json"
	"github.com/pkg/errors"
	"time"
)

// PaymentIntent is the base model for all payment intents.
type PaymentIntent struct {
	StripeIntentId            string    `db:"intent_id" json:"intent_id"`
	Origin                    string    `db:"origin" json:"origin"`
	Status                    string    `db:"status" json:"status"`
	Currency                  string    `db:"currency" json:"currency"`
	OriginAmount              int64     `db:"origin_amount" json:"origin_amount"`
	CreatedAt                 *time.Time `db:"created_at" json:"created_at"`
	LastUpdate                *time.Time `db:"last_update" json:"last_update"`
	RequiresAction            bool      `json:"requires_action"`
	PaymentIntentClientSecret *string   `json:"client_secret"`
}

func (p *PaymentIntent) Patch(body []byte) error {
	// Step 1 - unmarshal json body into a map
	updates := map[string]interface{}{}
	err := json.Unmarshal(body, &updates)
	if err != nil {
		return errors.Wrap(err, "unmarshaling patch")
	}

	// Step 2 - verify if one of the fields that are present is read-only
	roFields := map[string]string{
		"intent_id":     "intent_id",
		"origin":        "origin",
		"origin_amount": "origin_amount",
		"currency":      "currency",
		"created_at":    "created_at",
		"last_update":   "last_update",
		"client_token":  "client_token",
	}
	for fld, label := range roFields {
		if _, ok := updates[fld]; ok {
			return errors.New("can't patch payment intent " + label)
		}
	}

	if err = stringField(&p.Status, updates, "status"); err != nil {
		return err
	}

	return err
}
