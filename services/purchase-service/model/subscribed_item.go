package model

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/veganbase/backend/chassis"
	it "github.com/veganbase/backend/services/item-service/model"
	"github.com/veganbase/backend/services/purchase-service/model/types"
	"time"
)

type SubscriptionItem struct {
	ID            string                   `db:"id" json:"id"`
	Owner         string                   `db:"owner" json:"owner"`
	ItemID        string                   `db:"item_id" json:"item_id"`
	ItemType      it.ItemType              `db:"item_type" json:"item_type"`
	AddressID     *string                  `db:"address_id" json:"address_id"`
	Origin        string                   `db:"origin" json:"origin"`
	Quantity      int                      `db:"quantity" json:"quantity"`
	OtherInfo     chassis.GenericMap       `db:"other_info" json:"other_info"`
	DeliveryEvery int                      `db:"delivery_every" json:"delivery_every"`
	Status        types.SubscriptionStatus `db:"status" json:"status"`
	NextDelivery  int                      `db:"next_delivery" json:"next_delivery"`
	LastDelivery  *time.Time               `db:"last_delivery" json:"last_delivery"`
	CreatedAt     time.Time                `db:"created_at" json:"created_at"`
	ActiveSince   *time.Time               `db:"active_since" json:"active_since"`
	PausedSince   *time.Time               `db:"paused_since" json:"paused_since"`
	DeletedAt     *time.Time               `db:"deleted_at" json:"deleted_at"`
}

func (si *SubscriptionItem) Patch(body []byte) error {
	// Step 1 - unmarshal json body into a map
	updates := map[string]interface{}{}
	err := json.Unmarshal(body, &updates)
	if err != nil {
		return errors.Wrap(err, "unmarshaling patch")
	}

	// Step 2 - verify if one of the fields that are present is read-only
	roFields := map[string]string{
		"id":            "id",
		"owner":         "owner",
		"item_id":       "item_id",
		"item_type":     "item_type",
		"status":        "status",
		"origin":        "origin",
		"next_delivery": "next_delivery",
		"created_at":    "created_at",
		"active_since":  "active_since",
		"last_delivery": "last_delivery",
		"paused_since":  "paused_since",
		"deleted_at":    "deleted_at",
	}

	for fld, label := range roFields {
		if _, ok := updates[fld]; ok {
			return errors.New("can't patch field " + label)
		}
	}

	if err = chassis.StringField(si.AddressID, updates, "address_id"); err != nil {
		return err
	}
	if err = chassis.IntField(&si.Quantity, updates, "quantity"); err != nil {
		return err
	}

	if err = chassis.IntField(&si.DeliveryEvery, updates, "delivery_every"); err != nil {
		return err
	}

	//recalculating next delivery
	si.NextDelivery = (int(si.LastDelivery.Month()) + si.DeliveryEvery) % 12

	// Step 4 - patch OtherInfo field.
	other := map[string]interface{}(si.OtherInfo)
	for k, v := range updates {
		other[k] = v
	}

	si.OtherInfo = other

	return err
}
