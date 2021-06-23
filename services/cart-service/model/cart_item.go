package model

import (
	"bytes"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/cart-service/model/types"
	"github.com/veganbase/backend/services/item-service/model"
	"strings"
	"strconv"
)

// CartItem is a row in the cart_items table.
type CartItem struct {
	CartItemFixed
	OtherInfo types.OtherInfo `db:"other_info" json:"-,omitempty"`
}

type CartItemFixed struct {
	ID            int            `db:"id" json:"id"`
	CartID        string         `db:"cart_id" json:"cart_id,omitempty"`
	ItemID        string         `db:"item_id" json:"item_id"`
	Quantity      int            `db:"quantity" json:"quantity"`
	Subscribe     bool           `db:"subscribe" json:"subscribe"`
	DeliveryEvery int            `db:"delivery_every" json:"delivery_every"`
	Type          model.ItemType `db:"item_type" json:"item_type"`
}

func (ci *CartItem) UnmarshalJSON(data []byte) error {
	// Step 1.
	fields := map[string]interface{}{}
	err := json.Unmarshal(data, &fields)
	if err != nil {
		return errors.New("invalid JSON in cart-item data")
	}

	*ci = CartItem{}
	chassis.IntField(&ci.ID, fields, "id")
	chassis.StringField(&ci.CartID, fields, "cart_id")
	chassis.StringField(&ci.ItemID, fields, "item_id")
	if err = chassis.IntField(&ci.Quantity, fields, "quantity"); err != nil {
		return err
	}

	if err = chassis.BoolField(&ci.Subscribe, fields, "subscribe"); err != nil {
		return err
	}

	if err = chassis.IntField(&ci.DeliveryEvery, fields, "delivery_every"); err != nil {
		return err
	}

	//getting item type from db column or request
	itemType := ""
	chassis.StringField(&itemType, fields, "item_type")
	if itemType != "" {
		if err = ci.Type.FromString(itemType); err != nil {
			return err
		}
	}

	//if item_type is not defined on request, infer it from the prefix of item_id.
	// we are using strings split because currently rooms prefix is inconsistent (2 chars long, instead of 3)
	if ci.Type == model.UnknownItem {
		prefix := strings.Split(ci.ItemID, "_")
		for k, v := range model.ItemTypeIDPrefixes {
			if v == prefix[0] {
				ci.Type = k
				break
			}
		}
		if ci.Type == model.UnknownItem {
			return errors.New("invalid item type for item '" + strconv.Itoa(ci.ID) + "'")
		}

	}

	// Step 4.

	otherInfoFields, err := json.Marshal(fields)
	if err != nil {
		return errors.Wrap(err, "marshalling other-info fields to JSON")
	}

	// Step 5.
	if ci.Type != model.ProductOfferingItem && ci.Type != model.DishItem {
		it := "cart-" + ci.Type.String()
		attrRes, err := Validate(it, otherInfoFields)
		if err != nil {
			return errors.Wrap(err, "unmarshalling cart-item attributes for '"+it+"'")
		}
		if !attrRes.Valid() {
			msgs := []string{}
			for _, err := range attrRes.Errors() {
				msgs = append(msgs, err.String())
			}
			return errors.New("validation errors in cart-item attributes for '" + it +
				"': " + strings.Join(msgs, "; "))
		}

		// Step 7.
		ci.OtherInfo = fields
	}

	return nil
}

func (ci *CartItem) MarshalJSON() ([]byte, error) {
	// Marshal fixed fields.
	jsonFixed, err := json.Marshal(ci.CartItemFixed)
	if err != nil {
		return nil, err
	}

	// Return right away if there are no attributes or links.
	if len(ci.OtherInfo) == 0 {
		return jsonFixed, nil
	}
	// Compose into final JSON.
	var b bytes.Buffer
	b.Write(jsonFixed[:len(jsonFixed)-1])
	if len(ci.OtherInfo) > 0 {
		// Marshall type-specific attributes.
		jsonOtherInfo, err := json.Marshal(ci.OtherInfo)
		if err != nil {
			return nil, err
		}
		b.WriteByte(',')
		b.Write(jsonOtherInfo[1 : len(jsonOtherInfo)-1])
	}

	b.WriteByte('}')
	return b.Bytes(), nil
}

func (ci *CartItem) Patch(body []byte, itemType model.ItemType) error {
	// Step 1 - unmarshal json body into a map
	updates := map[string]interface{}{}
	err := json.Unmarshal(body, &updates)
	if err != nil {
		return errors.Wrap(err, "unmarshaling patch")
	}

	// Step 2 - verify if one of the fields that are present is read-only
	roFields := map[string]string{
		"id":        "id",
		"cart_id":   "cart_id",
		"item_id":   "item_id",
		"item_type": "item_type",
	}

	for fld, label := range roFields {
		if _, ok := updates[fld]; ok {
			return errors.New("can't patch item " + label)
		}
	}

	// Step 3 - update int field values
	if err = chassis.IntField(&ci.Quantity, updates, "quantity"); err != nil {
		return err
	}

	ci.Type = itemType

	other := map[string]interface{}(ci.OtherInfo)

	// Step 5.
	for k, v := range updates {
		other[k] = v
	}

	// Step 6.
	otherInfoFields, err := json.Marshal(other)
	if err != nil {
		return errors.New("couldn't marshal patched attributes back to JSON")
	}

	// Step 7.
	if ci.Type != model.ProductOfferingItem && ci.Type != model.DishItem && len(otherInfoFields) > 0 {
		itt := "cart-" + ci.Type.String()
		result, err := Validate(itt, otherInfoFields)
		if err != nil {
			return errors.Wrap(err, "processing item attributes for '"+itt+"'")
		}
		if !result.Valid() {
			msgs := []string{}
			for _, err := range result.Errors() {
				msgs = append(msgs, err.String())
			}
			return errors.New("validation errors in item attributes for '" + itt +
				"': " + strings.Join(msgs, "; "))
		}
	}

	ci.OtherInfo = other

	return err
}
