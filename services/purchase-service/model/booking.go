package model

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/veganbase/backend/services/purchase-service/model/types"
	"time"
)

// Booking is a row in the bookings table.
type Booking struct {
	Id            string              `db:"id" json:"id"`
	Origin        string              `db:"origin" json:"origin"`
	Host          string              `db:"host" json:"host"`
	BuyerID       string              `db:"buyer_id" json:"buyer_id"`
	ItemID        string              `db:"item_id" json:"item_id"`
	PaymentStatus types.PaymentStatus `db:"payment_status" json:"payment_status"`
	BookingInfo   types.BookingInfo   `db:"booking_info" json:"booking_info"`
	CreatedAt     time.Time           `db:"created_at" json:"created_at"`
	//OtherStatus   types.OtherMap      `db:"other_status"`
}

func (bk *Booking) Patch(body []byte) error {
	// Step 1 - unmarshal json body into a map
	updates := map[string]interface{}{}
	err := json.Unmarshal(body, &updates)
	if err != nil {
		return errors.Wrap(err, "unmarshaling patch")
	}

	// Step 2 - verify if one of the fields that are present is read-only
	roFields := map[string]string{
		"id":       "id",
		"origin":   "origin",
		"host":     "host",
		"buyer_id": "buyer_id",
	}

	for fld, label := range roFields {
		if _, ok := updates[fld]; ok {
			return errors.New("can't patch booking " + label)
		}
	}

	// Step 3 - update payment_status values
	var tempPaymentStatus string
	if err = stringField(&tempPaymentStatus, updates, "payment_status"); err != nil {
		return err
	}
	err = bk.PaymentStatus.FromString(tempPaymentStatus)

	var bkInfo map[string]interface{}

	bytes, _ := json.Marshal(bk.BookingInfo)
	if err = json.Unmarshal(bytes, &bkInfo); err != nil {
		return err
	}

	bookingInfo := updates["booking_info"]
	if bookingInfo != nil {
		newInfo := bookingInfo.(map[string]interface{})

		// Step 5.
		for k, v := range newInfo {
			bkInfo[k] = v
		}

		var jsonBytes []byte
		if jsonBytes, err = json.Marshal(bkInfo); err != nil {
			return err
		}

		if err = json.Unmarshal(jsonBytes, &bk.BookingInfo); err != nil {
			return err
		}
	}
	return err
}
