package model

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/cart-service/model/types"
	"time"
)

// Cart is the base model for all carts.
type Cart struct {
	// Unique ID of the cart.
	ID string `db:"id" json:"id"`

	// Status type, e.g. "active", "completed", "abandoned", "not logged in"
	CartStatus types.CartStatus `db:"cart_status" json:"cart_status"`

	// User ID of the registered owner of this item.
	Owner string `db:"owner" json:"owner"`

	// Creation time of item.
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

func (ct *Cart) Patch(body []byte) error {
	// Step 1 - unmarshal json body into a map
	updates := map[string]interface{}{}
	err := json.Unmarshal(body, &updates)
	if err != nil {
		return errors.Wrap(err, "unmarshaling patch")
	}

	// Step 2 - verify if one of the fields that are present is read-only
	roFields := map[string]string{
		"id":         "id",
		"created_at": "created_at",
	}
	for fld, label := range roFields {
		if _, ok := updates[fld]; ok {
			return errors.New("can't patch item " + label)
		}
	}

	// Step 3 - update field values
	if err = chassis.StringField(&ct.Owner, updates, "owner"); err != nil {
		return err
	}
	var tempCartStatus string
	if err = chassis.StringField(&tempCartStatus, updates, "cart_status"); err != nil {
		return err
	}
	err = ct.CartStatus.FromString(tempCartStatus)

	return err
}
