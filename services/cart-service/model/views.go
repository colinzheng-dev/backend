package model

import (
	"bytes"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/veganbase/backend/chassis"
)

type FullCart struct {
	Cart
	Items        []CartItemFull `json:"items"`
	DeliveryFees []DeliveryFee  `json:"delivery_fees"`
	IsValid      bool           `json:"is_valid"`
}

func FullView(cart *Cart, items *[]CartItem, deliveryFee *[]DeliveryFee, errors map[int][]string) *FullCart {
	view := FullCart{}
	fullCIs := []CartItemFull{}
	cartIsValid := true
	for _, ci := range *items {
		cif := CartItemFull{CartItem: ci, HasIssues: false, Errors: []string{}}
		if val, ok := errors[ci.ID]; ok {
			if len(val) > 0 {
				cif.HasIssues = true
				cif.Errors = val
				cartIsValid = false
			}
		}
		fullCIs = append(fullCIs, cif)
	}
	view.Cart = *cart
	view.Items = fullCIs
	view.DeliveryFees = *deliveryFee
	view.IsValid = cartIsValid
	return &view
}

type DeliveryFee struct {
	Seller            string `json:"seller"`
	Price             int    `json:"price"`
	FreeDeliveryAbove int    `json:"free_delivery_above"`
	Currency          string `json:"currency"`
}

type AvailabilityZone struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Reference int    `json:"reference"`
}

type CartItemFull struct {
	CartItem
	HasIssues bool     `json:"has_issues"`
	Errors    []string `json:"errors"`
}

func (cif *CartItemFull) MarshalJSON() ([]byte, error) {
	// Marshal fixed fields.
	jsonFixed, err := json.Marshal(cif.CartItemFixed)
	if err != nil {
		return nil, err
	}

	// Return right away if there are no attributes or links.
	isInvalid, err := json.Marshal(cif.HasIssues)
	if err != nil {
		return nil, err
	}
	ciErrors, err := json.Marshal(cif.Errors)
	if err != nil {
		return nil, err
	}
	// Compose into final JSON.
	var b bytes.Buffer

	b.Write(jsonFixed[:len(jsonFixed)-1])
	b.Write([]byte(`, "has_issues": `))
	b.Write(isInvalid)
	b.Write([]byte(`, "errors": `))
	b.Write(ciErrors)
	if len(cif.OtherInfo) > 0 {
		// Marshall type-specific attributes.
		jsonOtherInfo, err := json.Marshal(cif.OtherInfo)
		if err != nil {
			return nil, err
		}
		b.WriteByte(',')
		b.Write(jsonOtherInfo[1 : len(jsonOtherInfo)-1])
	}

	b.WriteByte('}')
	return b.Bytes(), nil
}


func (cif *CartItemFull) UnmarshalJSON(data []byte) error {
	// Step 1.
	fields := map[string]interface{}{}
	err := json.Unmarshal(data, &fields)
	if err != nil {
		return errors.New("invalid JSON in full-cart data")
	}

	*cif = CartItemFull{}

	chassis.IntField(&cif.ID, fields, "id")
	chassis.StringField(&cif.CartID, fields, "cart_id")
	chassis.StringField(&cif.ItemID, fields, "item_id")
	chassis.IntField(&cif.Quantity, fields, "quantity")
	chassis.BoolField(&cif.Subscribe, fields, "subscribe")
	chassis.IntField(&cif.DeliveryEvery, fields, "delivery_every")

	var itType string
	chassis.StringField(&itType, fields, "item_type")
	if err = cif.Type.FromString(itType); err != nil {
		return err
	}

	chassis.BoolField(&cif.HasIssues, fields, "has_issues")

	chassis.StringSliceField(&cif.Errors, fields, "errors")

	cif.OtherInfo = fields

	return nil
}

func (fc *FullCart) UnmarshalJSON(data []byte) error {
	// Step 1.
	fields := map[string]interface{}{}
	err := json.Unmarshal(data, &fields)
	if err != nil {
		return errors.New("invalid JSON in full-cart data")
	}

	*fc = FullCart{}

	chassis.StringField(&fc.ID, fields, "id")
	var status string
	chassis.StringField(&status, fields, "cart_status")
	if err = fc.CartStatus.FromString(status); err != nil {
		return err
	}
	chassis.TimeField(&fc.CreatedAt, fields, "created_at")
	chassis.BoolField(&fc.IsValid, fields, "is_valid")
	chassis.StringField(&fc.Owner, fields, "owner")

	items := fields["items"]

	rawItems, err := json.Marshal(items)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(rawItems, &fc.Items); err != nil {
		return err
	}

	fees := fields["delivery_fees"]
	rawFees, err := json.Marshal(fees)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(rawFees, &fc.DeliveryFees); err != nil {
		return err
	}

	return nil
}

