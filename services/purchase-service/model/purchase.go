package model

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/veganbase/backend/chassis"
	cartUtils "github.com/veganbase/backend/services/cart-service/server"
	"github.com/veganbase/backend/services/item-service/model"
	"github.com/veganbase/backend/services/purchase-service/model/types"
	"strings"
	"time"
)

// Purchase is the base model for all purchases.
type Purchase struct {
	Id            string               `db:"id" json:"id"`
	Status        types.PurchaseStatus `db:"status" json:"status"`
	BuyerID       string               `db:"buyer_id" json:"buyer_id"`
	Items         types.PurchaseItems  `db:"items" json:"items"`
	DeliveryFees  *DeliveryFees        `db:"delivery_fees" json:"delivery_fees,omitempty"`
	Site          *string              `db:"site" json:"site"`
	PaymentMethod string               `json:"payment_method"`
	CreatedAt     time.Time            `db:"created_at" json:"created_at"`
}

func (p *Purchase) Patch(body []byte) error {
	// Step 1 - unmarshal json body into a map
	updates := map[string]interface{}{}
	err := json.Unmarshal(body, &updates)
	if err != nil {
		return errors.Wrap(err, "unmarshaling patch")
	}

	// Step 2 - verify if one of the fields that are present is read-only
	roFields := map[string]string{
		"id":            "id",
		"created_at":    "created_at",
		"items":         "items",
		"buyer_id":      "buyer_id",
		"site":          "site",
		"delivery_fees": "delivery_fees",
	}
	for fld, label := range roFields {
		if _, ok := updates[fld]; ok {
			return errors.New("can't patch item " + label)
		}
	}
	//MAYBE WE CAN CHANGE THE ITEMS IN THE FUTURE?
	var tempPurchaseStatus string
	if err = stringField(&tempPurchaseStatus, updates, "status"); err != nil {
		return err
	}
	err = p.Status.FromString(tempPurchaseStatus)

	return err
}

type PurchaseRequest struct {
	AddressId string `json:"address,omitempty"`
}

type SimplePurchaseRequest struct {
	SimplePurchaseRequestFixed
	OtherInfo chassis.GenericMap `json:"other_info,omitempty"`
}

type SimplePurchaseRequestFixed struct {
	Name            string         `json:"name"`
	Email           string         `json:"email"`
	PaymentMethodID string         `json:"payment_method_id"`
	ItemID          string         `json:"item_id"`
	ItemType        model.ItemType `json:"-"`
	Quantity        int            `json:"quantity"`
}

func (sp *SimplePurchaseRequest) UnmarshalJSON(data []byte) error {
	// Step 1.
	fields := map[string]interface{}{}
	err := json.Unmarshal(data, &fields)
	if err != nil {
		return errors.New("invalid JSON in cart-item data")
	}

	// Step 2 - validate fixed fields against schema
	res, err := Validate("simple-purchase-fixed", data)
	if err != nil {
		return errors.Wrap(err, "unmarshalling fixed simple-purchase fields")
	}
	if !res.Valid() {
		msgs := []string{}
		for _, err := range res.Errors() {
			msgs = append(msgs, err.String())
		}
		return errors.New("validation errors for fixed simple-purchase fields: " + strings.Join(msgs, "; "))
	}

	*sp = SimplePurchaseRequest{}

	stringField(&sp.Name, fields, "name")
	stringField(&sp.Email, fields, "email")
	stringField(&sp.ItemID, fields, "item_id")
	stringField(&sp.PaymentMethodID, fields, "payment_method_id")
	intField(&sp.Quantity, fields, "quantity")

	//inferring item_type based on prefix
	prefix := strings.Split(sp.ItemID, "_")
	for k, v := range model.ItemTypeIDPrefixes {
		if v == prefix[0] {
			sp.ItemType = k
			break
		}
	}
	//validate itemType
	if sp.ItemType != model.RoomItem && sp.ItemType != model.OfferItem {
		return errors.New("invalid item for simple-purchase '" + sp.ItemType.String() + "'")
	}

	// Step 4.
	otherInfoFields, err := json.Marshal(fields)
	if err != nil {
		return errors.Wrap(err, "marshalling attributes fields to JSON")
	}
	it := "sp-" + sp.ItemType.String()
	attrRes, err := Validate(it, otherInfoFields)
	if err != nil {
		return errors.Wrap(err, "unmarshalling simple-purchase attributes for '"+it+"'")
	}
	if !attrRes.Valid() {
		msgs := []string{}
		for _, err := range attrRes.Errors() {
			msgs = append(msgs, err.String())
		}
		return errors.New("validation errors in simple-purchase attributes for '" + it +
			"': " + strings.Join(msgs, "; "))
	}

	// Step 5.

	switch sp.ItemType {
	case model.OfferItem:
		//fixing quantity in case it was omitted at the request
		sp.Quantity = 1
		period := fields["period"].(map[string]interface{})
		if len(period) < 0 {
			start, end := period["start"].(string), period["end"].(string)
			if err = cartUtils.ValidatePeriod(start, end); err != nil {
				return err
			}
		}
		//checking if start-time is valid
		if _, err = cartUtils.ValidateDatetime(fields["time-start"].(string)); err != nil {
			return err
		}
	case model.RoomItem:
		period := fields["period"].(map[string]interface{})
		start, end := period["start"].(string), period["end"].(string)
		if err = cartUtils.ValidatePeriod(start, end); err != nil {
			return err
		}
		//getting quantity of days based on the difference between end and start date
		if sp.Quantity, err = cartUtils.GetNumberOfDays(start, end); err != nil {
			return err
		}
	}

	sp.OtherInfo = fields

	return nil
}
