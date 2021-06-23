package model

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/veganbase/backend/services/purchase-service/model/types"
	"time"
)

// Order is a row in the orders table.
type Order struct {
	Id            string              `db:"id" json:"id"`
	Origin        string              `db:"origin" json:"origin"`
	BuyerID       string              `db:"buyer_id" json:"buyer_id"`
	Seller        string              `db:"seller" json:"seller"`
	PaymentStatus types.PaymentStatus `db:"payment_status" json:"payment_status"`
	Items         types.PurchaseItems `db:"items" json:"items"`
	DeliveryFee   *DeliveryFee        `db:"delivery_fee" json:"delivery_fee,omitempty"`
	OrderInfo     types.InfoMap       `db:"order_info" json:"order_info,omitempty"`
	CreatedAt     time.Time           `db:"created_at" json:"created_at"`
}

func (o *Order) Patch(body []byte) error {
	// Step 1 - unmarshal json body into a map
	updates := map[string]interface{}{}
	err := json.Unmarshal(body, &updates)
	if err != nil {
		return errors.Wrap(err, "unmarshaling patch")
	}

	// Step 2 - verify if one of the fields that are present is read-only
	roFields := map[string]string{
		"id":         "id",
		"code":       "code",
		"origin":     "origin",
		"buyer_id":   "buyer_id",
		"seller":     "seller",
		"items":      "items",
		"created_at": "created_at",
	}

	for fld, label := range roFields {
		if _, ok := updates[fld]; ok {
			return errors.New("can't patch item " + label)
		}
	}

	// Step 3 - update payment_status values
	var tempPaymentStatus string
	if err = stringField(&tempPaymentStatus, updates, "payment_status"); err != nil {
		return err
	}
	err = o.PaymentStatus.FromString(tempPaymentStatus)

	// Step 4 - patch MoreInfo field.
	other := map[string]interface{}(o.OrderInfo)
	for k, v := range updates {
		other[k] = v
	}

	o.OrderInfo = other

	return err
}
